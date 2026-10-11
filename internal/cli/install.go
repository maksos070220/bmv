package cli

import (
	"fmt"
	"time"

	"github.com/maksos070220/bmv/internal/app"
	"github.com/maksos070220/bmv/internal/browser"
	"github.com/maksos070220/bmv/internal/platform"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <browser>[@<version>]",
	Short: "Install a browser version",
	Example: `bmv install chrome
  bmv install chrome@120`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browserName, browserVersion, err := parseBrowserVersion(args[0])
		if err != nil {
			return err
		}

		b := browser.Browser(browserName)

		switch b {
		case browser.Chrome:
			// Supported.
		default:
			return fmt.Errorf("unsupported browser: %s", browserName)
		}

		currentPlatform, err := platform.Current()
		if err != nil {
			return err
		}

		application, err := app.New()
		if err != nil {
			return err
		}

		start := time.Now()
		out := cmd.OutOrStdout()

		release, err := application.BrowserService.Install(
			cmd.Context(),
			b,
			browserVersion,
			currentPlatform,
			func(downloaded, total int64) {
				renderProgress(out, start, downloaded, total)
			},
		)

		if err != nil {
			return fmt.Errorf("install browser: %w", err)
		}

		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}

		_, err = fmt.Fprintf(out, "Installed %s %s\n", release.Browser, release.Version)

		return err
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
