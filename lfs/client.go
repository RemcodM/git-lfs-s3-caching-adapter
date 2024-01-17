package lfs

import (
	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tq"
)

type LFSTransferClient struct {
	config    *config.Configuration
	lfsClient *lfsapi.Client
	manifest  tq.Manifest
	operation string
	remote    string
}

func NewLFSTransferClient(cfg *config.Configuration, operation string, remote string) (*LFSTransferClient, error) {
	client, err := getAPIClient(cfg)
	if err != nil {
		return nil, err
	}
	return &LFSTransferClient{
		config:    cfg,
		lfsClient: client,
		manifest:  getManifest(cfg, client, operation, remote),
		operation: operation,
		remote:    remote,
	}, nil
}

func (c *LFSTransferClient) NewTransferQueue(progressCallback tools.CopyCallback) *tq.TransferQueue {
	tqOperation := tq.Upload
	if c.operation == "download" {
		tqOperation = tq.Download
	}
	return tq.NewTransferQueue(
		tqOperation,
		c.manifest,
		c.config.Remote(),
		tq.RemoteRef(currentRemoteRef(c.config, c.remote)),
		tq.WithBatchSize(1),
		tq.WithProgressCallback(progressCallback),
	)
}

func (c *LFSTransferClient) Close() error {
	return c.lfsClient.Close()
}

func (c *LFSTransferClient) Operation() string {
	return c.operation
}

func (c *LFSTransferClient) IsDownload() bool {
	return c.operation == "download"
}

func (c *LFSTransferClient) IsUpload() bool {
	return c.operation == "upload"
}

func (c *LFSTransferClient) Remote() string {
	return c.remote
}

func getManifest(cfg *config.Configuration, client *lfsapi.Client, operation string, remote string) tq.Manifest {
	return tq.NewManifest(cfg.Filesystem(), client, operation, remote)
}

func getAPIClient(cfg *config.Configuration) (*lfsapi.Client, error) {
	return lfsapi.NewClient(cfg)
}

func currentRemoteRef(cfg *config.Configuration, remote string) *git.Ref {
	return git.NewRefUpdate(cfg.Git, remote, cfg.CurrentRef(), nil).RemoteRef()
}
