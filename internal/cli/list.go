package cli

import (
	"fmt"

	"github.com/maksos070220/bmv/internal/app"
	"github.com/maksos070220/bmv/internal/browser"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list <browser>",
	Short:   "List installed browser versions",
	Example: `  bmv list chrome`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b := browser.Browser(args[0])

		switch b {
		case browser.Chrome:
		default:
			return fmt.Errorf("unsupported browser: %s", b)
		}

		application, err := app.New()
		if err != nil {
			return fmt.Errorf("initialize application: %w", err)
		}

		versions, err := application.BrowserService.ListInstalledVersions(b)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "No installed versions of %s\n", b)
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed versions of %s:\n", b)
		if err != nil {
			return err
		}

		for _, v := range versions {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", v); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
