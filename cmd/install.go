/*
Copyright © 2024 Remco de Man <remco@heliumnet.nl>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/v3/git"
	gitlfs "github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/spf13/cobra"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/lfs"
)

var (
	fileInstall     = ""
	forceInstall    = false
	localInstall    = false
	systemInstall   = false
	worktreeInstall = false
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Git LFS S3 caching adapter configuration.",
	Long: `Set-ups the Git LFS S3 caching adapter as a standalone custom transfer adapter
for Git LFS.
	
It will only register as a standalone custom transfer adapter for 'caching::'
Git LFS urls in the users .gitconfig. This can be overriden using the various
flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		destArgs := 0
		if localInstall {
			destArgs++
		}
		if worktreeInstall {
			destArgs++
		}
		if systemInstall {
			destArgs++
		}
		if fileInstall != "" {
			destArgs++
		}

		if destArgs > 1 {
			fmt.Fprintln(os.Stderr, "Only one of the --local, --system, --worktree, and --file options can be specified.")
			os.Exit(1)
		}

		uid := os.Geteuid()
		if systemInstall && uid != 0 && uid != -1 {
			fmt.Fprintln(os.Stderr, "warning: current user is not root/admin, system install is likely to fail.")
		}

		options := gitlfs.FilterOptions{
			GitConfig:  nil,
			Force:      forceInstall,
			File:       fileInstall,
			Local:      localInstall,
			Worktree:   worktreeInstall,
			System:     systemInstall,
			SkipSmudge: false,
		}
		if err := lfs.InstallAttributes(&options); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s\n", err.Error())
			fmt.Fprintln(os.Stderr, "Run `git lfs install --force` to reset Git configuration.")
			os.Exit(2)
		}
		fmt.Fprintf(os.Stdout, "Git LFS S3 caching adapter initialized.\n")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS S3 caching adapter config, overwriting previous values.")
	installCmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS S3 caching adapter config for the local Git repository only.")
	installCmd.Flags().StringVarP(&fileInstall, "file", "", "", "Set the Git LFS S3 caching adapter config for the given configuration file only.")
	installCmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Set the Git LFS S3 caching adapter config in system-wide scope.")
	if git.IsGitVersionAtLeast("2.20.0") {
		installCmd.Flags().BoolVarP(&worktreeInstall, "worktree", "w", false, "Set the Git LFS S3 caching adapter config for the current Git working tree, if multiple working trees are configured; otherwise, the same as --local.")
	}
}
