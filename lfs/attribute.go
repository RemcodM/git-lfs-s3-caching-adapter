package lfs

import (
	"errors"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
)

func changeToWorkingCopy(config *config.Configuration) error {
	workingDir := config.LocalWorkingDir()
	cwd, err := tools.Getwd()
	if err != nil {
		return errors.New("could not determine current working directory")
	}
	cwd, err = tools.CanonicalizeSystemPath(cwd)
	if err != nil {
		return errors.New("could not canonicalize current working directory")
	}

	// If the current working directory is not within the repository's
	// working directory, then let's change directories accordingly.  This
	// should only occur if GIT_WORK_TREE is set.
	if !(strings.HasPrefix(cwd, workingDir) && (cwd == workingDir || (len(cwd) > len(workingDir) && cwd[len(workingDir)] == os.PathSeparator))) {
		os.Chdir(workingDir)
	}
	return nil
}

func checkOptions(options *lfs.FilterOptions) error {
	config := GetPassthroughConfiguration()
	if options.Local || options.Worktree {
		if !config.InRepo() {
			return errors.New("not in a git repository")
		}
		bare, err := git.IsBare()
		if err != nil {
			return errors.New("could not determine bareness")
		}
		if !bare {
			err = changeToWorkingCopy(config)
			if err != nil {
				return err
			}
		}
	}
	options.GitConfig = config.GitConfig()
	return nil
}

func InstallAttributes(options *lfs.FilterOptions) error {
	err := checkOptions(options)
	if err != nil {
		return err
	}

	cachingTransferAgentAttribute().Install(options)
	standaloneCachingTransferAgentAttribute().Install(options)
	return nil
}

func UninstallAttributes(options *lfs.FilterOptions) error {
	err := checkOptions(options)
	if err != nil {
		return err
	}

	cachingTransferAgentAttribute().Uninstall(options)
	standaloneCachingTransferAgentAttribute().Uninstall(options)
	return nil
}

func cachingTransferAgentAttribute() *lfs.Attribute {
	executable, err := os.Executable()
	if err != nil {
		executable = "git-lfs-s3-caching-adapter"
	}
	return &lfs.Attribute{
		Section: "lfs.customtransfer.caching",
		Properties: map[string]string{
			"path": executable,
		},
		Upgradeables: map[string][]string{
			"path": {
				"git-lfs-s3-caching-adapter",
			},
		},
	}
}

func standaloneCachingTransferAgentAttribute() *lfs.Attribute {
	return &lfs.Attribute{
		Section: "lfs.caching::",
		Properties: map[string]string{
			"standalonetransferagent": "caching",
		},
		Upgradeables: map[string][]string{
			"standalonetransferagent": {
				"caching",
			},
		},
	}
}
