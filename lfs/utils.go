package lfs

import (
	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tq"
)

func GetPassthroughConfiguration() *config.Configuration {
	config := config.New()
	config.Git = newPassthroughEnvironment(config.Git)
	return config
}

func GetTransferQueue(cfg *config.Configuration, operation string, remote string, progressCallback tools.CopyCallback) (*tq.TransferQueue, *lfsapi.Client, error) {
	tqOperation := tq.Upload
	if operation == "download" {
		tqOperation = tq.Download
	}
	manifest, apiClient, err := getTransferManifestOperationRemote(cfg, operation, remote)
	if err != nil {
		return nil, nil, err
	}
	return tq.NewTransferQueue(
		tqOperation,
		manifest,
		cfg.Remote(),
		tq.RemoteRef(currentRemoteRef(cfg, remote)),
		tq.WithBatchSize(1),
		tq.WithProgressCallback(progressCallback),
	), apiClient, nil
}

func currentRemoteRef(cfg *config.Configuration, remote string) *git.Ref {
	return git.NewRefUpdate(cfg.Git, remote, cfg.CurrentRef(), nil).RemoteRef()
}

func getTransferManifestOperationRemote(cfg *config.Configuration, operation string, remote string) (tq.Manifest, *lfsapi.Client, error) {
	c, err := getAPIClient(cfg)
	if err != nil {
		return nil, nil, err
	}

	return tq.NewManifest(cfg.Filesystem(), c, operation, remote), c, nil
}

func getAPIClient(cfg *config.Configuration) (*lfsapi.Client, error) {
	return lfsapi.NewClient(cfg)
}
