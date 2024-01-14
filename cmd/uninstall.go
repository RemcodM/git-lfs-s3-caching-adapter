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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Git LFS S3 caching adapter by removing transfer adapter configuration.",
	Long: `Removes the custom transfer adapter configuration for the Git LFS S3 caching
adapter from the Git configuration`,
	Run: func(cmd *cobra.Command, args []string) {
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
		if err := lfs.UninstallAttributes(&options); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s\n", err.Error())
			os.Exit(2)
		}
		if systemInstall {
			fmt.Fprintf(os.Stdout, "System Git LFS S3 caching adapter configuration has been removed.\n")
		} else if !(localInstall || worktreeInstall) {
			fmt.Fprintf(os.Stdout, "Global Git LFS S3 caching adapter configuration has been removed.\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Remove the Git LFS S3 caching adapter config for the local Git repository only.")
	uninstallCmd.Flags().StringVarP(&fileInstall, "file", "", "", "Remove the Git LFS S3 caching adapter config for the given configuration file only.")
	uninstallCmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Remove the Git LFS S3 caching adapter config in system-wide scope.")
	if git.IsGitVersionAtLeast("2.20.0") {
		uninstallCmd.Flags().BoolVarP(&worktreeInstall, "worktree", "w", false, "Remove the Git LFS S3 caching adapter config for the current Git working tree, if multiple working trees are configured; otherwise, the same as --local.")
	}
}
