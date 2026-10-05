package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/maksos070220/bmv/internal/platform"
	"github.com/maksos070220/bmv/internal/providers/chrome"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <browser>@<version>",
	Short: "Resolve a browser version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browserName, browserVersion, err := parseBrowserVersion(args[0])
		if err != nil {
			return err
		}

		if browserName != "chrome" {
			return fmt.Errorf("unsupported browser: %s", browserName)
		}

		currentPlatform, err := platform.Current()
		if err != nil {
			return err
		}

		provider := chrome.NewProvider(nil)

		release, err := provider.Resolve(context.Background(), browserVersion, currentPlatform)
		if err != nil {
			return err
		}

		fmt.Printf("Browser:  %s\n", release.Browser)
		fmt.Printf("Version:  %s\n", release.Version)
		fmt.Printf("Platform: %s/%s\n", release.Platform.OS, release.Platform.Architecture)
		fmt.Printf("Download: %s\n", release.DownloadURL)

		return nil
	},
}

func parseBrowserVersion(value string) (string, string, error) {
	parts := strings.SplitN(value, "@", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid browser version %q, expected <browser>@<version>", value)
	}

	return parts[0], parts[1], nil
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}
