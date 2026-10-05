package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maksos070220/bmv/internal/browser"
)

const defaultLatestVersionsURL = "https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json"

type Provider struct {
	client            *http.Client
	latestVersionsURL string
}

func NewProvider(client *http.Client) *Provider {
	if client == nil {
		client = http.DefaultClient
	}

	return &Provider{
		client:            client,
		latestVersionsURL: defaultLatestVersionsURL,
	}
}

func (p *Provider) Browser() browser.Browser {
	return browser.Chrome
}

type manifest struct {
	Milestones map[string]milestone `json:"milestones"`
}

type milestone struct {
	Version   string    `json:"version"`
	Revision  string    `json:"revision"`
	Downloads downloads `json:"downloads"`
}

type downloads struct {
	Chrome []download `json:"chrome"`
}

type download struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

func (p *Provider) Resolve(ctx context.Context, version string, platform browser.Platform) (*browser.Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.latestVersionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Chrome manifest: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chrome manifest returned HTTP %d", resp.StatusCode)
	}

	var data manifest

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode Chrome manifest: %w", err)
	}

	m, ok := data.Milestones[version]
	if !ok {
		return nil, fmt.Errorf(
			"Chrome milestone %q not found",
			version,
		)
	}

	cfTPlatform, err := platformName(platform)
	if err != nil {
		return nil, err
	}

	for _, item := range m.Downloads.Chrome {
		if item.Platform != cfTPlatform {
			continue
		}

		return &browser.Release{
			Browser:     browser.Chrome,
			Version:     m.Version,
			Platform:    platform,
			DownloadURL: item.URL,
		}, nil
	}

	return nil, fmt.Errorf("chrome %s is not available for %s/%s", version, platform.OS, platform.Architecture)
}

func platformName(platform browser.Platform) (string, error) {
	switch {
	case platform.OS == "linux" && platform.Architecture == "amd64":
		return "linux64", nil

	case platform.OS == "linux" && platform.Architecture == "arm64":
		return "linux-arm64", nil

	case platform.OS == "darwin" && platform.Architecture == "amd64":
		return "mac-x64", nil

	case platform.OS == "darwin" && platform.Architecture == "arm64":
		return "mac-arm64", nil

	case platform.OS == "windows" && platform.Architecture == "386":
		return "win32", nil

	case platform.OS == "windows" && platform.Architecture == "amd64":
		return "win64", nil

	default:
		return "", fmt.Errorf(
			"unsupported platform: %s/%s", platform.OS, platform.Architecture)
	}
}

var _ browser.Provider = (*Provider)(nil)
