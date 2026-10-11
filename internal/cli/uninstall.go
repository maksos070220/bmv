package cli

import (
	"fmt"
	"regexp"

	"github.com/maksos070220/bmv/internal/app"
	"github.com/maksos070220/bmv/internal/browser"
	"github.com/spf13/cobra"
)

var chromeVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <browser>@<version>",
	Short:   "Uninstall an installed browser version",
	Example: "  bmv uninstall chrome@120.0.6099.109",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browserName, browserVersion, err := parseBrowserVersion(args[0])
		if err != nil {
			return err
		}

		b := browser.Browser(browserName)

		switch b {
		case browser.Chrome:
		default:
			return fmt.Errorf("unsupported browser: %s", browserName)
		}

		if !chromeVersionPattern.MatchString(browserVersion) {
			return fmt.Errorf("invalid Chrome version %q: expected a full version such as 120.0.6099.109", browserVersion)
		}

		application, err := app.New()
		if err != nil {
			return fmt.Errorf("initialize application: %w", err)
		}

		if err := application.BrowserService.Uninstall(b, browserVersion); err != nil {
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled %s %s\n", b, browserVersion)

		return err
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
