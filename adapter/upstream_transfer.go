package adapter

import (
	"github.com/git-lfs/git-lfs/v3/tq"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/lfs"
)

type upstreamTransfer struct {
	completedTransfer *tq.Transfer
	oid               string
	path              string
	size              int64
	transferQueue     *tq.TransferQueue
	OnProgress        func(oid string, totalSize int64, readSoFar int64, readSinceLast int64) error
	OnFinished        func(oid string, path string, size int64, result *tq.Transfer)
}

func NewUpstreamTransfer(client *lfs.LFSTransferClient, oid string, path string, size int64) *upstreamTransfer {
	transfer := &upstreamTransfer{
		completedTransfer: nil,
		oid:               oid,
		path:              path,
		size:              size,
		OnProgress:        nil,
		OnFinished:        nil,
	}
	transfer.transferQueue = client.NewTransferQueue(func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if transfer.OnProgress != nil {
			if err := transfer.OnProgress(transfer.oid, totalSize, readSoFar, int64(readSinceLast)); err != nil {
				return err
			}
		}
		return nil
	})
	return transfer
}

func (u *upstreamTransfer) Perform() {
	go func() {
		for transfer := range u.transferQueue.Watch() {
			u.completedTransfer = transfer
		}
	}()
	u.transferQueue.Add(u.oid, u.path, u.oid, u.size, false, nil)
	go func() {
		u.transferQueue.Wait()
		if u.OnFinished != nil {
			u.OnFinished(u.oid, u.path, u.size, u.completedTransfer)
		}
	}()
}
