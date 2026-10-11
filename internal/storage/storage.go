package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/maksos070220/bmv/internal/browser"
)

type Storage struct {
	root string
}

func New(root string) *Storage {
	return &Storage{
		root: root,
	}
}

func Default() (*Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home directory: %w", err)
	}

	return New(filepath.Join(home, ".bmv")), nil
}

func (s *Storage) Root() string {
	return s.root
}

func (s *Storage) BrowserPath(b browser.Browser, version string) string {
	return filepath.Join(s.root, "browsers", string(b), version)
}

func (s *Storage) CachePath() string {
	return filepath.Join(s.root, "cache")
}

func (s *Storage) DownloadsPath() string {
	return filepath.Join(s.root, "downloads")
}

func (s *Storage) IsInstalled(b browser.Browser, version string) bool {
	path := s.BrowserPath(b, version)

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
