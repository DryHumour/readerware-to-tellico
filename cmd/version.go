package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time variables set via ldflags
var (
	version = "v0.0.0-dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Prints the version, commit, build date, and builder information.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, cmd.Root().Name(), "version", version)
		fmt.Fprintln(out, "commit:", commit)
		fmt.Fprintln(out, "built at:", date)
		fmt.Fprintln(out, "built by:", builtBy)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
