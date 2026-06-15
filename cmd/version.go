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
		fmt.Println(cmd.Root().Name(), "version", version)
		fmt.Println("commit:", commit)
		fmt.Println("built at:", date)
		fmt.Println("built by:", builtBy)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
