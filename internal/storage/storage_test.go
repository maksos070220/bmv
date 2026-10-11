package storage

import (
	"path/filepath"
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
