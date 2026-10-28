package cli

import (
	"fmt"

	"github.com/maksos070220/bmv/internal/app"
	"github.com/maksos070220/bmv/internal/browser"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run <browser>@<version> -- [browser flags]",
	Short:              "Run an installed browser version",
	Example:            "  bmv run chrome@120.0.6099.109 -- --headless",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		browserName, versionBrowser, err := parseBrowserVersion(args[0])
		if err != nil {
			return err
		}

		b := browser.Browser(browserName)
		if b != browser.Chrome {
			return fmt.Errorf("unsupported browser: %s", browserName)
		}

		application, err := app.New()
		if err != nil {
			return fmt.Errorf("initialize application: %w", err)
		}

		browserArgs := args[1:]
		if len(browserArgs) > 0 && browserArgs[0] == "--" {
			browserArgs = browserArgs[1:]
		}

		if err := application.BrowserService.Run(cmd.Context(), b, versionBrowser, browserArgs); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
