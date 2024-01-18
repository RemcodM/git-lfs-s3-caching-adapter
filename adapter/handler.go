package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/v3/tq"

	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/caching"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/lfs"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/stats"
)

type cachingHandler struct {
	cacheAdapter *caching.S3CachingAdapter
	client       *lfs.LFSTransferClient
	output       *os.File
	stats        *stats.Stats
	tempdir      string
}

// newHandler creates a new handler for the protocol.
func newHandler(output *os.File, msg *inputMessage) (*cachingHandler, error) {
	config := lfs.GetPassthroughConfiguration()

	tempdir, err := os.MkdirTemp(config.Filesystem().LFSStorageDir, "lfs-caching-adapter-*")
	if err != nil {
		return nil, err
	}

	client, err := lfs.NewLFSTransferClient(config, msg.Operation, msg.Remote)
	if err != nil {
		return nil, err
	}

	cacheAdapter, err := caching.NewS3CachingAdapter(config)
	if err != nil {
		return nil, err
	}

	return &cachingHandler{
		cacheAdapter: cacheAdapter,
		client:       client,
		output:       output,
		stats:        stats.NewSessionStats(),
		tempdir:      tempdir,
	}, nil
}

func (h *cachingHandler) onUpstreamFinished(oid string, path string, size int64, result *tq.Transfer) {
	if result == nil {
		if h.client.IsUpload() {
			fmt.Fprintf(os.Stderr, "Expected upload of object %s, but no action was performed. Most likely, the object already exists remotely. Continuing...\n", oid)
		} else {
			fmt.Fprintf(os.Stderr, "Expected download of object %s, but upstream transfer did not perform any action. Returning error.\n", oid)
			h.complete(oid, path, fmt.Errorf("no action performed, but expected download of file %s", oid))
			return
		}
	} else {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Got a transfer result error for %s: %s\n", result.Oid, result.Error.Error())
			h.complete(result.Oid, result.Path, errors.New(result.Error.Error()))
			return
		}
		oid = result.Oid
		path = result.Path
		size = result.Size
	}

	if h.client.IsDownload() {
		h.stats.ObjectsPulled++
		h.stats.BytesTransferredFromRemote += uint64(size)
	} else {
		h.stats.ObjectsPushed++
		h.stats.BytesTransferredToRemote += uint64(size)
	}

	if h.cacheAdapter != nil {
		fmt.Fprintf(os.Stderr, "Adding object %s to cache\n", oid)
		uploaded, err := h.cacheAdapter.Upload(path, oid, size)
		if uploaded {
			if h.client.IsDownload() {
				h.stats.CacheAddedDuringPull++
			} else {
				h.stats.CacheAddedDuringPush++
			}
			h.stats.BytesTransferredToCache += uint64(size)
			fmt.Fprintf(os.Stderr, "Added object %s to cache\n", oid)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error while adding object %s to cache. %s Object is not cached for next download.\n", oid, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Object %s is already in cache\n", oid)
		}
	}

	h.complete(oid, path, nil)
}

func (h *cachingHandler) onProgress(oid string, totalSize int64, bytesSoFar int64, bytesSinceLast int64) error {
	response := &progressMessage{
		Event:          "progress",
		Oid:            oid,
		BytesSoFar:     bytesSoFar,
		BytesSinceLast: bytesSinceLast,
	}
	json.NewEncoder(h.output).Encode(response)
	return nil
}

func (h *cachingHandler) upstreamTransfer(oid string, path string, size int64) {
	transfer := NewUpstreamTransfer(h.client, oid, path, size)
	transfer.OnFinished = h.onUpstreamFinished
	transfer.OnProgress = h.onProgress
	transfer.Perform()
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
		h.terminate()
		return false
	default:
		standaloneFailure(fmt.Sprintf("unknown event %q", msg.Event), nil)
	}
	return true
}

// upload performs the upload action for the given OID, size, and path.
func (h *cachingHandler) upload(oid string, size int64, path string) {
	fmt.Fprintf(os.Stderr, "Passing upload of object %s to upstream adapter (path: %s)\n", oid, path)
	h.upstreamTransfer(oid, path, size)
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
			h.onProgress(oid, size, bytesSoFar, bytesSinceLast)
		})
		if ok {
			h.stats.ObjectsPulled++
			h.stats.CacheHits++
			h.stats.BytesTransferredFromCache += uint64(size)
			fmt.Fprintf(os.Stderr, "Downloaded object %s from cache to target %s\n", oid, tmp.Name())
			h.complete(oid, tmp.Name(), err)
			return
		} else if err == nil {
			h.stats.CacheMisses++
			fmt.Fprintf(os.Stderr, "Cache miss for object %s. Will download upstream instead.\n", oid)
		} else {
			h.stats.CacheErrors++
			fmt.Fprintf(os.Stderr, "Cache error while obtaining object %s. %s Will download upstream instead.\n", oid, err.Error())
		}
	}

	fmt.Fprintf(os.Stderr, "Passing uncached download of object %s to upstream adapter, target: %s\n", oid, tmp.Name())
	h.upstreamTransfer(oid, tmp.Name(), size)
}

func (h *cachingHandler) terminate() error {
	fmt.Fprintf(os.Stderr, "Received call to terminate, writing stats\n")
	err := h.stats.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed writing stats, ignoring...\n")
	}
	return h.client.Close()
}
