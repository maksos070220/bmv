package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print BMV version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bmv %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
