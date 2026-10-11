package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bmv",
	Short: "Browser Version Manager",
	Long: `BMV is a version manager for browsers.

Install, manage and run multiple browser versions locally.`,
}

func Execute(ctx context.Context) error {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		_, err = fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}
