package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maksos070220/bmv/internal/browser"
)

func TestStorage_BrowserPath(t *testing.T) {
	root := t.TempDir()

	store := New(root)

	got := store.BrowserPath(browser.Chrome, "120.0.6099.109")

	expected := filepath.Join(root, "browsers", "chrome", "120.0.6099.109")

	if got != expected {
		t.Fatalf("expected path %q, got %q", expected, got)
	}
}

func TestStorage_CachePath(t *testing.T) {
	root := t.TempDir()

	store := New(root)

	got := store.CachePath()

	expected := filepath.Join(root, "cache")

	if got != expected {
		t.Fatalf("expected path %q, got %q", expected, got)
	}
}

func TestStorage_DownloadsPath(t *testing.T) {
	root := t.TempDir()

	store := New(root)

	got := store.DownloadsPath()

	expected := filepath.Join(root, "downloads")

	if got != expected {
		t.Fatalf("expected path %q, got %q", expected, got)
	}
}

func TestStorage_ListInstalledVersions(t *testing.T) {
	root := t.TempDir()
	storage := New(root)

	versions := []string{
		"120.0.6099.109",
		"99.0.4844.51",
		"131.0.6778.204",
	}

	for _, version := range versions {
		path := storage.BrowserPath(browser.Chrome, version)

		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", path, err)
		}
	}

	browserDir := filepath.Join(root, "browsers", string(browser.Chrome))
	if err := os.WriteFile(filepath.Join(browserDir, "README"), []byte("not a browser"), 0644); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}

	got, err := storage.ListInstalledVersions(browser.Chrome)
	if err != nil {
		t.Fatalf("ListInstalledVersions() error = %v", err)
	}

	want := []string{
		"120.0.6099.109",
		"131.0.6778.204",
		"99.0.4844.51",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListInstalledVersions() = %v, want %v", got, want)
	}
}

func TestStorage_ListInstalledVersionsEmpty(t *testing.T) {
	storage := New(t.TempDir())

	got, err := storage.ListInstalledVersions(browser.Chrome)
	if err != nil {
		t.Fatalf("ListInstalledVersions() error = %v", err)
	}

	if got == nil {
		t.Fatal("ListInstalledVersions() returned nil, want empty slice")
	}

	if len(got) != 0 {
		t.Errorf("ListInstalledVersions() = %v, want empty slice", got)
	}
}

func TestStorage_ListInstalledVersionsIgnoresFiles(t *testing.T) {
	storage := New(t.TempDir())

	browserDir := filepath.Join(storage.Root(), "browsers", string(browser.Chrome))

	if err := os.MkdirAll(browserDir, 0755); err != nil {
		t.Fatalf("MkdirAll(): %v", err)
	}

	if err := os.WriteFile(filepath.Join(browserDir, "not-a-version"), []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}

	got, err := storage.ListInstalledVersions(browser.Chrome)
	if err != nil {
		t.Fatalf("ListInstalledVersions() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("ListInstalledVersions() = %v, want empty slice", got)
	}
}
