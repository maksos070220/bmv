package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func (s *Storage) ListInstalledVersions(b browser.Browser) ([]string, error) {
	dir := filepath.Join(s.root, "browsers", string(b))

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("read installed browser versions: %w", err)
	}

	versions := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versions = append(versions, entry.Name())
	}

	sort.Strings(versions)

	return versions, nil
}

func (s *Storage) RemoveInstalledVersion(b browser.Browser, version string) error {
	if version == "" || version == "." || version == ".." || strings.ContainsAny(version, `/\`) {
		return fmt.Errorf("invalid browser version %q", version)
	}

	path := s.BrowserPath(b, version)

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("browser %s version %s is not installed", b, version)
		}

		return fmt.Errorf("inspect browser installation: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("browser installation path is a symbolic link")
	}

	if !info.IsDir() {
		return fmt.Errorf("browser installation path is not a directory")
	}

	if err = os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove browser installation: %w", err)
	}

	return nil
}
