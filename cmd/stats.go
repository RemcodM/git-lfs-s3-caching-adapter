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
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/stats"
)

var (
	jsonStats  = false
	totalStats = false
	siUnits    = false
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics collected for the current repository",
	Long: `Show various statistics collected for the current repository by the various
sessions of the caching adapter. These can be used to gain insight in the amount
of bandwidth saved by the cache, but also as a way to guarantee the cache is
working correctly.

Please note that when concurrency for the caching adapter is enabled in Git LFS, one
LFS command will result in multiple sessions of the caching adapter, thus resulting
in multiple statistics entries.`,
	Run: func(cmd *cobra.Command, args []string) {
		allStats, errs := stats.ReadAllStats()
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err.Error())
			}
			cmd.PrintErrf("warning: some statistics could not be read\n\n")
		}

		unmarkedStats, _ := compact(cmd, stats.FilterUnmarked(allStats), true)

		var outputStats *stats.Stats
		var err error
		if totalStats {
			outputStats, err = stats.CollectedStats(allStats)
		} else {
			outputStats, err = stats.CollectedStats(unmarkedStats)
		}
		if err != nil {
			cmd.PrintErrln(err.Error())
			cmd.PrintErrf("warning: could not collect statistics\n")
			os.Exit(1)
		}

		if jsonStats {
			json, err := json.Marshal(outputStats)
			if err != nil {
				cmd.PrintErrln(err.Error())
				cmd.PrintErrf("warning: could not encode statistics as JSON\n")
				os.Exit(1)
			}
			cmd.Println(string(json))
		} else {
			byteFormatFunc := stats.ByteCountIEC
			if siUnits {
				byteFormatFunc = stats.ByteCountSI
			}
			if totalStats {
				cmd.Printf("Collected statistics for %d sessions since repository clone:\n\n", outputStats.Sessions)
			} else {
				cmd.Printf("Collected statistics for %d sessions:\n\n", outputStats.Sessions)
			}
			cmd.Printf("Objects pulled:                %d\n", outputStats.ObjectsPulled)
			cmd.Printf("  Cache hits:                  %d (%s)\n", outputStats.CacheHits, stats.Percentage(outputStats.CacheHits, outputStats.ObjectsPulled))
			cmd.Printf("  Cache misses:                %d (%s)\n", outputStats.CacheMisses, stats.Percentage(outputStats.CacheMisses, outputStats.ObjectsPulled))
			cmd.Printf("  Cache errors:                %d (%s)\n", outputStats.CacheErrors, stats.Percentage(outputStats.CacheErrors, outputStats.ObjectsPulled))
			cmd.Printf("  Cache additions during pull: %d (%s)\n\n", outputStats.CacheAddedDuringPull, stats.Percentage(outputStats.CacheAddedDuringPull, outputStats.ObjectsPulled))

			cmd.Printf("Objects pushed:                %d\n", outputStats.ObjectsPushed)
			cmd.Printf("  Cache additions during push: %d (%s)\n\n", outputStats.CacheAddedDuringPush, stats.Percentage(outputStats.CacheAddedDuringPush, outputStats.ObjectsPushed))

			cmd.Printf("Bytes downloaded from remote:  %s\n", byteFormatFunc(outputStats.BytesTransferredFromRemote))
			cmd.Printf("Bytes downloaded from cache:   %s\n", byteFormatFunc(outputStats.BytesTransferredFromCache))
			cmd.Printf("Bytes uploaded to remote:      %s\n", byteFormatFunc(outputStats.BytesTransferredToRemote))
			cmd.Printf("Bytes uploaded to cache:       %s\n", byteFormatFunc(outputStats.BytesTransferredToCache))
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().BoolVarP(&skipCompact, "no-compact", "C", false, "Do not automatically compact statistics")
	statsCmd.Flags().BoolVarP(&jsonStats, "json", "j", false, "Use machine readable JSON output format for the statistics")
	statsCmd.Flags().BoolVarP(&siUnits, "si", "s", false, "Use SI units when printing statistics (e.g. 1000 bytes = 1kb), instead of IEC units. This only affects the human readable output format")
	statsCmd.Flags().BoolVarP(&totalStats, "total", "t", false, "Export total statistics for this repository, not only the statistics since last reset")
}
