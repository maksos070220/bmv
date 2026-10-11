package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/maksos070220/bmv/internal/browser"
	"github.com/maksos070220/bmv/internal/downloader"
)

type Downloader interface {
	Download(ctx context.Context, url string, destination string, onProgress downloader.ProgressFunc) error
}

type Installer interface {
	Install(archivePath string, destination string) error
}

type Storage interface {
	BrowserPath(b browser.Browser, version string) string
	DownloadsPath() string
	IsInstalled(browser browser.Browser, version string) bool
	ListInstalledVersions(b browser.Browser) ([]string, error)
}

type ProviderRegistry interface {
	Get(browser.Browser) (browser.Provider, error)
}

type BrowserService struct {
	providers  ProviderRegistry
	downloader Downloader
	installer  Installer
	storage    Storage
}

func NewBrowserService(providers ProviderRegistry, downloader Downloader, installer Installer, storage Storage) *BrowserService {
	return &BrowserService{
		providers:  providers,
		downloader: downloader,
		installer:  installer,
		storage:    storage,
	}
}

func (s *BrowserService) Resolve(ctx context.Context, b browser.Browser, version string, platform browser.Platform) (*browser.Release, error) {
	provider, err := s.providers.Get(b)
	if err != nil {
		return nil, err
	}

	return provider.Resolve(ctx, version, platform)
}

func (s *BrowserService) Install(ctx context.Context, b browser.Browser, version string, platform browser.Platform, onProgress func(downloaded, total int64)) (*browser.Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("install browser: %w", err)
	}

	release, err := s.Resolve(ctx, b, version, platform)
	if err != nil {
		return nil, fmt.Errorf("resolve browser: %w", err)
	}

	if s.storage.IsInstalled(release.Browser, release.Version) {
		return release, nil
	}

	archivePath := filepath.Join(s.storage.DownloadsPath(), fmt.Sprintf("%s-%s.zip", release.Browser, release.Version))

	if err = s.downloader.Download(ctx, release.DownloadURL, archivePath, onProgress); err != nil {
		return nil, fmt.Errorf("download browser: %w", err)
	}

	if err = ctx.Err(); err != nil {
		return nil, fmt.Errorf("install browser: %w", err)
	}

	destination := s.storage.BrowserPath(release.Browser, release.Version)

	if err = s.installer.Install(archivePath, destination); err != nil {
		return nil, fmt.Errorf("install browser: %w", err)
	}

	return release, nil
}

func (s *BrowserService) ListInstalledVersions(b browser.Browser) ([]string, error) {
	versions, err := s.storage.ListInstalledVersions(b)
	if err != nil {
		return nil, fmt.Errorf("list installed browser versions: %w", err)
	}

	return versions, nil
}
