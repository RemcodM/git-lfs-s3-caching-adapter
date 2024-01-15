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

	"github.com/spf13/cobra"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/adapter"
)

var rootCmd = &cobra.Command{
	Use:   "git-lfs-s3-caching-adapter",
	Short: "Standalone transfer adapter which caches Git LFS objects in a S3 bucket",
	Long: `Git LFS S3 caching adapter is a standalone transfer adapter which caches Git
LFS objects in a S3 bucket.

It uses the underlying Git LFS implementation to perform transfers using the
LFS configuration you set-up for your repository, but additionally caches the
downloaded/uploaded objects in a S3 bucket you configure. This way, it is
possbile to cache large objects at a closer edge location, possibly reducing
download time and bandwidth costs.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := adapter.ProcessData(os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(2)
		}
	},
}

func Execute() {
	rootCmd.SetOut(os.Stdout)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
