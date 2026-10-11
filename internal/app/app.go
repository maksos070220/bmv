package app

import (
	"net/http"

	"github.com/maksos070220/bmv/internal/downloader"
	"github.com/maksos070220/bmv/internal/installer"
	"github.com/maksos070220/bmv/internal/providers/chrome"
	"github.com/maksos070220/bmv/internal/registry"
	"github.com/maksos070220/bmv/internal/service"
	"github.com/maksos070220/bmv/internal/storage"
)

type App struct {
	BrowserService *service.BrowserService
}

func New() (*App, error) {
	store, err := storage.Default()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}

	chromeProvider := chrome.NewProvider(httpClient)

	providers := registry.New(chromeProvider)

	browserService := service.NewBrowserService(providers, downloader.New(httpClient), installer.New(), store)

	return &App{
		BrowserService: browserService,
	}, nil
}
