package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maksos070220/bmv/internal/browser"
)

const defaultLastKnownGoodURL = "https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json"
const defaultLatestVersionsURL = "https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json"

type Provider struct {
	client            *http.Client
	lastKnownGoodURL  string
	latestVersionsURL string
}

func NewProvider(client *http.Client) *Provider {
	if client == nil {
		client = http.DefaultClient
	}

	return &Provider{
		client:            client,
		lastKnownGoodURL:  defaultLastKnownGoodURL,
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

type lastKnownGoodManifest struct {
	Channels channels `json:"channels"`
}

type channels struct {
	Stable channel `json:"Stable"`
}

type channel struct {
	Channel   string    `json:"channel"`
	Version   string    `json:"version"`
	Revision  string    `json:"revision"`
	Downloads downloads `json:"downloads"`
}

func (p *Provider) Resolve(ctx context.Context, version string, platform browser.Platform) (*browser.Release, error) {
	cfTPlatform, err := platformName(platform)
	if err != nil {
		return nil, err
	}

	if version == "latest" || version == "" {
		return p.resolveLatest(ctx, platform, cfTPlatform)
	}

	return p.resolveMilestone(ctx, version, platform, cfTPlatform)
}

func (p *Provider) getResponse(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Chrome manifest: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("chrome manifest returned HTTP %d", resp.StatusCode)
	}

	return resp, nil
}

func (p *Provider) resolveMilestone(ctx context.Context, version string, platform browser.Platform, cfTPlatform string) (*browser.Release, error) {
	resp, err := p.getResponse(ctx, p.latestVersionsURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var data manifest

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode Chrome manifest: %w", err)
	}

	browserVersion, ok := data.Milestones[version]
	if !ok {
		return nil, fmt.Errorf("chrome milestone %q not found", version)
	}

	for _, item := range browserVersion.Downloads.Chrome {
		if item.Platform != cfTPlatform {
			continue
		}

		return &browser.Release{
			Browser:     browser.Chrome,
			Version:     browserVersion.Version,
			Platform:    platform,
			DownloadURL: item.URL,
		}, nil
	}

	return nil, fmt.Errorf("chrome %s is not available for %s/%s", version, platform.OS, platform.Architecture)
}

func (p *Provider) resolveLatest(ctx context.Context, platform browser.Platform, cfTPlatform string) (*browser.Release, error) {
	resp, err := p.getResponse(ctx, p.lastKnownGoodURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var data lastKnownGoodManifest

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode Chrome manifest: %w", err)
	}

	for _, item := range data.Channels.Stable.Downloads.Chrome {
		if item.Platform != cfTPlatform {
			continue
		}

		return &browser.Release{
			Browser:     browser.Chrome,
			Version:     data.Channels.Stable.Version,
			Platform:    platform,
			DownloadURL: item.URL,
		}, nil
	}

	return nil, fmt.Errorf("chrome stable is not available for %s/%s", platform.OS, platform.Architecture)
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
