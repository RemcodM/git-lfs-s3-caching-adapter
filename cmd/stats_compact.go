/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/stats"
)

var (
	skipCompact  = false
	skipMarked   = false
	skipUnmarked = false
)

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact the collected statistics for the current repository",
	Long: `Statistics are internally stored as one JSON file per session. Depending on the
Git LFS concurrency settings and the number of executed commands, this can
result in a large number of files. This command compacts all statistics into a
single file, such that statistics performance is drastically improved.`,
	Run: func(cmd *cobra.Command, args []string) {
		code := 0
		allStats, errs := stats.ReadAllStats()
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err.Error())
			}
			cmd.PrintErrf("warning: some statistics could not be read\n\n")
			code = 1
		}

		if !skipMarked {
			if compact, ok := compact(cmd, stats.FilterMarked(allStats), false); ok {
				if len(compact) > 0 && verbose {
					cmd.Printf("Compacted %d post-reset sessions\n", compact[0].Sessions)
				}
			} else {
				code = 1
			}
		}

		if !skipUnmarked {
			if compact, ok := compact(cmd, stats.FilterUnmarked(allStats), false); ok {
				if len(compact) > 0 && verbose {
					cmd.Printf("Compacted %d pre-reset sessions\n", compact[0].Sessions)
				}
			} else {
				code = 1
			}
		}
		os.Exit(code)
	},
}

func compact(cmd *cobra.Command, compactableStats []stats.Stats, auto bool) ([]stats.Stats, bool) {
	if auto && (skipCompact || len(compactableStats) < 20) {
		return compactableStats, true
	}
	if auto {
		cmd.PrintErrf("info: automatically compacting %d statistics objects\n", len(compactableStats))
	}
	collectedStats, errs := stats.Compact(compactableStats)
	if len(errs) > 0 {
		for _, err := range errs {
			cmd.PrintErrln(err.Error())
		}
		cmd.PrintErrf("warning: there were errors while compacting\n\n")
	}
	if collectedStats != nil && !collectedStats.IsZero() {
		return []stats.Stats{*collectedStats}, len(errs) == 0
	}
	return []stats.Stats{}, len(errs) == 0
}

func init() {
	statsCmd.AddCommand(compactCmd)

	compactCmd.Flags().BoolVarP(&skipMarked, "skip-marked", "M", false, "Do not remove marked, pre-reset statistics")
	compactCmd.Flags().BoolVarP(&skipUnmarked, "skip-unmarked", "U", false, "Do not remove unmarked, post-reset statistics")
}
