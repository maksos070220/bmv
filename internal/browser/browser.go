package browser

import "context"

type Browser string

const (
	Chrome  Browser = "chrome"
	Firefox Browser = "firefox"
	Edge    Browser = "edge"
)

type Platform struct {
	OS           string
	Architecture string
}

type Release struct {
	Browser     Browser
	Version     string
	Platform    Platform
	DownloadURL string
}

type Provider interface {
	Browser() Browser
	Resolve(ctx context.Context, version string, platform Platform) (*Release, error)
}
