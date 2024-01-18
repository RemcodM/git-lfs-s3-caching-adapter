/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/stats"
)

var purgeStats bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the statistics collected for the current repository",
	Long: `This resets the statistics collected for the current repository. By default, the
statistics are not really removed from the filesystem, but merely marked as
reset. This allows the user to still inspect the total collected statistics at
later time. This behaviour can be overridden using the --purge flag. When this
flag is set, the statistics are actually removed from the filesystem.`,
	Run: func(cmd *cobra.Command, args []string) {
		if purgeStats {
			err := stats.PurgeAll()
			if err != nil {
				cmd.PrintErrln(err.Error())
			}
			os.Exit(1)
		}

		code := 0
		allStats, errs := stats.ReadAllStats()
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err.Error())
			}
			cmd.PrintErrln("warning: some statistics could not be read")
		}

		compactedStats, ok := compact(cmd, allStats, true)
		if !ok {
			code = 1
		}

		if len(compactedStats) == 0 {
			return
		}
		errs = stats.MarkAll(compactedStats)
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err.Error())
			}
			cmd.PrintErrln("warning: some statistics could not be reset")
		}
		if verbose {
			cmd.Printf("Marked %d statistics object(s) as reset\n", len(compactedStats))
		}
		os.Exit(code)
	},
}

func init() {
	statsCmd.AddCommand(resetCmd)

	resetCmd.Flags().BoolVarP(&skipCompact, "no-compact", "C", false, "Do not automatically compact statistics during reset")
	resetCmd.Flags().BoolVarP(&purgeStats, "purge", "p", false, "Actually remove all statistics, instead of marking them as reset")
}
