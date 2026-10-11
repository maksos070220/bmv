package platform

import (
	"fmt"
	"runtime"

	"github.com/maksos070220/bmv/internal/browser"
)

func Current() (browser.Platform, error) {
	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		}, nil

	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return browser.Platform{
			OS:           "linux",
			Architecture: "arm64",
		}, nil

	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return browser.Platform{
			OS:           "darwin",
			Architecture: "amd64",
		}, nil

	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return browser.Platform{
			OS:           "darwin",
			Architecture: "arm64",
		}, nil

	case runtime.GOOS == "windows" && runtime.GOARCH == "386":
		return browser.Platform{
			OS:           "windows",
			Architecture: "386",
		}, nil

	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return browser.Platform{
			OS:           "windows",
			Architecture: "amd64",
		}, nil

	default:
		return browser.Platform{}, fmt.Errorf(
			"unsupported platform: %s/%s",
			runtime.GOOS,
			runtime.GOARCH,
		)
	}
}
