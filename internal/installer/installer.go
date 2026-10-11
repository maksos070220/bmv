package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Installer struct{}

func New() *Installer {
	return &Installer{}
}

func (i *Installer) Install(archivePath string, destination string) error {
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check destination: %w", err)
	}

	parent := filepath.Dir(destination)

	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("create installation directory: %w", err)
	}

	tempDir, err := os.MkdirTemp(parent, ".bmv-install-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if err = extractZip(archivePath, tempDir); err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	if err = os.Rename(tempDir, destination); err != nil {
		return fmt.Errorf("finalize installation: %w", err)
	}

	return nil
}

func extractZip(archivePath string, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}

	defer func() {
		_ = reader.Close()
	}()

	for _, file := range reader.File {
		if !filepath.IsLocal(file.Name) {
			return fmt.Errorf("archive contains unsafe path: %q", file.Name)
		}

		targetPath := filepath.Join(destination, file.Name)

		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("create directory %q: %w", file.Name, err)
			}

			continue
		}

		if err = os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("create parent directory for %q: %w", file.Name, err)
		}

		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archive entry %q: %w", file.Name, err)
		}

		dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode().Perm())
		if err != nil {
			_ = src.Close()

			return fmt.Errorf("create file %q: %w", file.Name, err)
		}

		_, copyErr := io.Copy(dst, src)
		closeDstErr := dst.Close()
		closeSrcErr := src.Close()

		if copyErr != nil {
			return fmt.Errorf("extract file %q: %w", file.Name, copyErr)
		}

		if closeDstErr != nil {
			return fmt.Errorf("close extracted file %q: %w", file.Name, closeDstErr)
		}

		if closeSrcErr != nil {
			return fmt.Errorf("close archive entry %q: %w", file.Name, closeSrcErr)
		}
	}

	return nil
}
