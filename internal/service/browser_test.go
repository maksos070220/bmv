package service

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maksos070220/bmv/internal/browser"
	"github.com/maksos070220/bmv/internal/downloader"
)

type mockProviderRegistry struct {
	provider browser.Provider
	err      error
}

func (m *mockProviderRegistry) Get(name browser.Browser) (browser.Provider, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.provider, nil
}

type mockProvider struct {
	release *browser.Release
	err     error
}

func (m *mockProvider) Browser() browser.Browser {
	return browser.Chrome
}

func (m *mockProvider) Resolve(ctx context.Context, version string, platform browser.Platform) (*browser.Release, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.release, nil
}

type mockDownloader struct {
	calls       int
	url         string
	destination string
	err         error
}

func (m *mockDownloader) Download(ctx context.Context, url string, destination string, onProgress downloader.ProgressFunc) error {
	m.calls++
	m.url = url
	m.destination = destination

	return m.err
}

type mockInstaller struct {
	calls       int
	archivePath string
	destination string
	err         error
}

func (m *mockInstaller) Install(archivePath string, destination string) error {
	m.calls++
	m.archivePath = archivePath
	m.destination = destination

	return m.err
}

type mockStorage struct {
	root           string
	installed      bool
	versions       []string
	listErr        error
	executablePath string
	executableErr  error
}

func (m *mockStorage) BrowserPath(b browser.Browser, version string) string {
	return filepath.Join(m.root, "browsers", string(b), version)
}

func (m *mockStorage) DownloadsPath() string {
	return filepath.Join(m.root, "downloads")
}

func (m *mockStorage) IsInstalled(b browser.Browser, version string) bool {
	return m.installed
}

func (m *mockStorage) ListInstalledVersions(b browser.Browser) ([]string, error) {
	return m.versions, m.listErr
}

func (m *mockStorage) RemoveInstalledVersion(b browser.Browser, version string) error {
	if m.listErr != nil {
		return m.listErr
	}

	return nil
}

func (m *mockStorage) ExecutablePath(b browser.Browser, version string) (string, error) {
	return m.executablePath, m.executableErr
}

type mockProcessRunner struct {
	calls      int
	executable string
	args       []string
	err        error
}

func (m *mockProcessRunner) Run(ctx context.Context, executable string, args []string) error {
	m.calls++
	m.executable = executable
	m.args = args

	return m.err
}

func TestBrowserService_Install(t *testing.T) {
	ctx := context.Background()

	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	provider := &mockProvider{
		release: release,
	}

	providers := &mockProviderRegistry{
		provider: provider,
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	got, err := svc.Install(ctx, browser.Chrome, "", browser.Platform{
		OS:           "linux",
		Architecture: "amd64",
	}, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if got != release {
		t.Fatalf("Install() release = %v, want %v", got, release)
	}

	if downloader.calls != 1 {
		t.Fatalf("Download() calls = %d, want 1", downloader.calls)
	}

	expectedArchive := filepath.Join(storage.DownloadsPath(), "chrome-120.0.1.2.zip")

	if downloader.url != release.DownloadURL {
		t.Errorf("Download() URL = %q, want %q", downloader.url, release.DownloadURL)
	}

	if downloader.destination != expectedArchive {
		t.Errorf("Download() destination = %q, want %q", downloader.destination, expectedArchive)
	}

	if installer.calls != 1 {
		t.Fatalf("Install() calls = %d, want 1", installer.calls)
	}

	if installer.archivePath != expectedArchive {
		t.Errorf("Installer archivePath = %q, want %q", installer.archivePath, expectedArchive)
	}

	expectedDestination := storage.BrowserPath(browser.Chrome, release.Version)

	if installer.destination != expectedDestination {
		t.Errorf("Installer destination = %q, want %q", installer.destination, expectedDestination)
	}
}

func TestBrowserService_InstallAlreadyInstalled(t *testing.T) {
	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	providers := &mockProviderRegistry{
		provider: &mockProvider{
			release: release,
		},
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root:      t.TempDir(),
		installed: true,
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	got, err := svc.Install(context.Background(), browser.Chrome, "", release.Platform, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if got != release {
		t.Fatalf("Install() release = %v, want %v", got, release)
	}

	if downloader.calls != 0 {
		t.Errorf("Download() calls = %d, want 0", downloader.calls)
	}

	if installer.calls != 0 {
		t.Errorf("Installer calls = %d, want 0", installer.calls)
	}
}

func TestBrowserService_InstallDownloadError(t *testing.T) {
	expectedErr := errors.New("network error")

	release := &browser.Release{
		Browser:     browser.Chrome,
		Version:     "120.0.1.2",
		DownloadURL: "https://example.com/chrome.zip",
	}

	providers := &mockProviderRegistry{
		provider: &mockProvider{
			release: release,
		},
	}

	downloader := &mockDownloader{
		err: expectedErr,
	}

	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	_, err := svc.Install(context.Background(), browser.Chrome, "",
		browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		}, nil,
	)

	if err == nil {
		t.Fatal("Install() error = nil, want error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Install() error = %v, want wrapped %v", err, expectedErr)
	}

	if downloader.calls != 1 {
		t.Errorf("Download() calls = %d, want 1", downloader.calls)
	}

	if installer.calls != 0 {
		t.Errorf("Installer calls = %d, want 0", installer.calls)
	}
}

func TestBrowserService_InstallInstallerError(t *testing.T) {
	expectedErr := errors.New("extraction failed")

	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	providers := &mockProviderRegistry{
		provider: &mockProvider{
			release: release,
		},
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{
		err: expectedErr,
	}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	_, err := svc.Install(context.Background(), browser.Chrome, "", release.Platform, nil)

	if err == nil {
		t.Fatal("Install() error = nil, want error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Install() error = %v, want wrapped %v", err, expectedErr)
	}

	if downloader.calls != 1 {
		t.Errorf("Download() calls = %d, want 1", downloader.calls)
	}

	if installer.calls != 1 {
		t.Errorf("Installer calls = %d, want 1", installer.calls)
	}
}

func TestBrowserService_InstallResolveError(t *testing.T) {
	expectedErr := errors.New("manifest unavailable")

	providers := &mockProviderRegistry{
		provider: &mockProvider{
			err: expectedErr,
		},
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	_, err := svc.Install(context.Background(), browser.Chrome, "",
		browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		}, nil,
	)

	if err == nil {
		t.Fatal("Install() error = nil, want error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Install() error = %v, want wrapped %v", err, expectedErr)
	}

	if downloader.calls != 0 {
		t.Errorf("Download() calls = %d, want 0", downloader.calls)
	}

	if installer.calls != 0 {
		t.Errorf("Installer calls = %d, want 0", installer.calls)
	}
}

func TestBrowserService_InstallUnsupportedBrowser(t *testing.T) {
	providers := &mockProviderRegistry{
		err: errors.New("unsupported browser: firefox"),
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	_, err := svc.Install(context.Background(), browser.Firefox, "", browser.Platform{
		OS:           "linux",
		Architecture: "amd64",
	}, nil,
	)

	if err == nil {
		t.Fatal("Install() error = nil, want error")
	}

	if downloader.calls != 0 {
		t.Errorf("Download() calls = %d, want 0", downloader.calls)
	}

	if installer.calls != 0 {
		t.Errorf("Installer calls = %d, want 0", installer.calls)
	}
}

func TestBrowserService_InstallCancelledContext(t *testing.T) {
	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	providers := &mockProviderRegistry{
		provider: &mockProvider{
			release: release,
		},
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Install(ctx, browser.Chrome, "", release.Platform, nil)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Install() error = %v, want %v", err, context.Canceled)
	}

	if downloader.calls != 0 {
		t.Errorf("Download() calls = %d, want 0", downloader.calls)
	}

	if installer.calls != 0 {
		t.Errorf("Installer calls = %d, want 0", installer.calls)
	}
}

func TestBrowserService_ListInstalledVersions(t *testing.T) {
	want := []string{
		"120.0.6099.109",
		"131.0.6778.204",
	}

	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	provider := &mockProvider{
		release: release,
	}

	providers := &mockProviderRegistry{
		provider: provider,
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root:     t.TempDir(),
		versions: want,
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	got, err := svc.ListInstalledVersions(browser.Chrome)
	if err != nil {
		t.Fatalf("ListInstalledVersions() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListInstalledVersions() = %v, want %v", got, want)
	}
}

func TestBrowserService_Uninstall(t *testing.T) {
	release := &browser.Release{
		Browser: browser.Chrome,
		Version: "120.0.1.2",
		Platform: browser.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		DownloadURL: "https://example.com/chrome.zip",
	}

	provider := &mockProvider{
		release: release,
	}

	providers := &mockProviderRegistry{
		provider: provider,
	}

	downloader := &mockDownloader{}
	installer := &mockInstaller{}
	storage := &mockStorage{
		root: t.TempDir(),
	}
	processRunner := &mockProcessRunner{}

	svc := NewBrowserService(providers, downloader, installer, storage, processRunner)

	err := svc.Uninstall(browser.Chrome, "120.0.6099.109")
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
}

func TestBrowserService_RunSuccess(t *testing.T) {
	expectedExecutable := "/tmp/chrome"
	expectedArgs := []string{
		"--headless",
		"--disable-gpu",
		"https://example.com",
	}

	storage := &mockStorage{
		executablePath: expectedExecutable,
	}
	runner := &mockProcessRunner{}

	svc := NewBrowserService(nil, nil, nil, storage, runner)

	err := svc.Run(context.Background(), browser.Chrome, "120.0.6099.109", expectedArgs)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	if runner.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", runner.calls)
	}

	if runner.executable != expectedExecutable {
		t.Errorf("executable = %q, want %q", runner.executable, expectedExecutable)
	}

	if !reflect.DeepEqual(runner.args, expectedArgs) {
		t.Errorf("args = %v, want %v", runner.args, expectedArgs)
	}
}

func TestBrowserService_RunExecutablePathError(t *testing.T) {
	expectedErr := errors.New("browser executable not found")

	storage := &mockStorage{
		executableErr: expectedErr,
	}
	runner := &mockProcessRunner{}

	svc := NewBrowserService(nil, nil, nil, storage, runner)

	err := svc.Run(context.Background(), browser.Chrome, "120.0.6099.109", nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("Run() error = %v, want wrapped %v", err, expectedErr)
	}

	if runner.calls != 0 {
		t.Errorf("runner calls = %d, want 0", runner.calls)
	}
}

func TestBrowserService_RunProcessError(t *testing.T) {
	expectedErr := errors.New("failed to start browser")

	storage := &mockStorage{
		executablePath: "/tmp/chrome",
	}
	runner := &mockProcessRunner{
		err: expectedErr,
	}

	svc := NewBrowserService(nil, nil, nil, storage, runner)

	err := svc.Run(context.Background(), browser.Chrome, "120.0.6099.109", nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("Run() error = %v, want wrapped %v", err, expectedErr)
	}

	if runner.calls != 1 {
		t.Errorf("runner calls = %d, want 1", runner.calls)
	}
}

func TestBrowserService_RunCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	storage := &mockStorage{
		executablePath: "/tmp/chrome",
	}
	runner := &mockProcessRunner{}

	svc := NewBrowserService(nil, nil, nil, storage, runner)

	err := svc.Run(ctx, browser.Chrome, "120.0.6099.109", nil)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}

	if runner.calls != 0 {
		t.Errorf("runner calls = %d, want 0", runner.calls)
	}
}
