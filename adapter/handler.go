package adapter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tq"

	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/caching"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/lfs"
)

type cachingHandler struct {
	apiClient     *lfsapi.Client
	cacheAdapter  *caching.S3CachingAdapter
	config        *config.Configuration
	currentOid    *string
	transferQueue *tq.TransferQueue
	output        *os.File
	tempdir       string
}

type cacheProgressHandler struct {
	handler *cachingHandler
}

// newHandler creates a new handler for the protocol.
func newHandler(output *os.File, msg *inputMessage) (*cachingHandler, error) {
	cfg := lfs.GetPassthroughConfiguration()
	progressHandler := &cacheProgressHandler{}
	transferQueue, apiClient, err := lfs.GetTransferQueue(cfg, msg.Operation, msg.Remote, func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if progressHandler.handler != nil && progressHandler.handler.currentOid != nil {
			progressHandler.handler.progress(*progressHandler.handler.currentOid, readSoFar, int64(readSinceLast))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	tempdir, err := os.MkdirTemp(cfg.Filesystem().LFSStorageDir, "lfs-caching-adapter-*")
	if err != nil {
		return nil, err
	}

	cacheAdapter, err := caching.NewS3CachingAdapter(cfg)
	if err != nil {
		return nil, err
	}

	handler := &cachingHandler{
		apiClient:     apiClient,
		cacheAdapter:  cacheAdapter,
		config:        cfg,
		transferQueue: transferQueue,
		output:        output,
		tempdir:       tempdir,
	}
	progressHandler.handler = handler

	transferWatch := transferQueue.Watch()

	go func() {
		for transfer := range transferWatch {
			handler.upstreamFinished(transfer)
		}
	}()

	return handler, nil
}

// dispatch dispatches the event depending on the message type.
func (h *cachingHandler) dispatch(msg *inputMessage) bool {
	switch msg.Event {
	case "init":
		fmt.Fprintf(os.Stderr, "Received initialization message\n")
		fmt.Fprintln(h.output, "{}")
	case "upload":
		h.upload(msg.Oid, msg.Size, msg.Path)
	case "download":
		h.download(msg.Oid, msg.Size)
	case "terminate":
		fmt.Fprintf(os.Stderr, "Received call to terminate\n")
		h.terminate()
		return false
	default:
		standaloneFailure(fmt.Sprintf("unknown event %q", msg.Event), nil)
	}
	return true
}

func (h *cachingHandler) upstreamFinished(transfer *tq.Transfer) {
	if transfer.Error != nil {
		h.complete(transfer.Oid, transfer.Path, transfer.Error)
		return
	}

	if h.cacheAdapter != nil {
		fmt.Fprintf(os.Stderr, "Adding object %s to cache\n", transfer.Oid)
		err := h.cacheAdapter.Upload(transfer.Path, transfer.Oid, transfer.Size)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while adding object %s to cache. %s Object is not cached for next download.\n", transfer.Oid, err.Error())
		}
	}

	h.complete(transfer.Oid, transfer.Path, nil)
}

func (h *cachingHandler) terminate() {
	h.apiClient.Close()
}

func (h *cachingHandler) progress(oid string, bytesSoFar int64, bytesSinceLast int64) {
	response := &progressMessage{
		Event:          "progress",
		Oid:            oid,
		BytesSoFar:     bytesSoFar,
		BytesSinceLast: bytesSinceLast,
	}
	json.NewEncoder(h.output).Encode(response)
}

// complete sends a response to an upload or download command, using the return
// values from those functions.
func (h *cachingHandler) complete(oid string, path string, err error) {
	response := &completeMessage{
		Event: "complete",
		Oid:   oid,
		Path:  path,
	}
	if err != nil {
		response.Error = &errorMessage{Message: err.Error()}
	}
	json.NewEncoder(h.output).Encode(response)
}

// upload performs the upload action for the given OID, size, and path.
func (h *cachingHandler) upload(oid string, size int64, path string) {
	fmt.Fprintf(os.Stderr, "Passing upload of object %s to upstream adapter (path: %s)\n", oid, path)
	h.currentOid = &oid
	h.transferQueue.Add(oid, path, oid, size, false, nil)
}

// download performs the download action for the given OID and size.
func (h *cachingHandler) download(oid string, size int64) {
	tmp, err := os.CreateTemp(h.tempdir, "download")
	if err != nil {
		h.complete(oid, "", err)
		return
	}
	tmp.Close()
	os.Remove(tmp.Name())

	if h.cacheAdapter != nil {
		fmt.Fprintf(os.Stderr, "Trying to download object %s from cache, target: %s\n", oid, tmp.Name())
		ok, err := h.cacheAdapter.Download(tmp.Name(), oid, size, func(bytesSoFar int64, bytesSinceLast int64) {
			h.progress(oid, bytesSoFar, bytesSinceLast)
		})
		if ok {
			fmt.Fprintf(os.Stderr, "Downloaded object %s from cache to target %s\n", oid, tmp.Name())
			h.complete(oid, tmp.Name(), err)
			return
		} else if err == nil {
			fmt.Fprintf(os.Stderr, "Cache miss for object %s. Will download upstream instead.\n", oid)
		} else {
			fmt.Fprintf(os.Stderr, "Cache error while obtaining object %s. %s Will download upstream instead.\n", oid, err.Error())
		}
	}

	fmt.Fprintf(os.Stderr, "Passing uncached download of object %s to upstream adapter, target: %s\n", oid, tmp.Name())
	h.currentOid = &oid
	h.transferQueue.Add(oid, tmp.Name(), oid, size, false, nil)
}
