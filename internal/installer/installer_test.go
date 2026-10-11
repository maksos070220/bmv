package installer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, entries map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "browser.zip")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)

	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			_ = file.Close()
			t.Fatalf("create zip entry: %v", err)
		}

		if _, err = entry.Write([]byte(content)); err != nil {
			_ = file.Close()
			t.Fatalf("write zip entry: %v", err)
		}
	}

	if err = writer.Close(); err != nil {
		_ = file.Close()
		t.Fatalf("close zip writer: %v", err)
	}

	if err = file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	return path
}

func TestInstaller_Install(t *testing.T) {
	archive := createTestZip(t, map[string]string{
		"chrome-linux64/chrome":         "chrome binary",
		"chrome-linux64/resources/data": "test data",
	})

	destination := filepath.Join(t.TempDir(), "chrome", "120.0.6099.109")

	installer := New()

	if err := installer.Install(archive, destination); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	chromePath := filepath.Join(destination, "chrome-linux64", "chrome")

	content, err := os.ReadFile(chromePath)
	if err != nil {
		t.Fatalf("read installed file: %v", err)
	}

	if string(content) != "chrome binary" {
		t.Fatalf("expected %q, got %q", "chrome binary", string(content))
	}
}

func TestInstaller_RejectsUnsafePath(t *testing.T) {
	archive := createTestZip(t, map[string]string{
		"../../evil": "malicious content",
	})

	destination := filepath.Join(t.TempDir(), "chrome", "120")

	installer := New()

	err := installer.Install(archive, destination)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination should not exist after failed installation")
	}
}

func TestInstaller_PreservesExecutablePermissions(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "browser.zip")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)

	header := &zip.FileHeader{
		Name: "chrome-linux64/chrome",
	}

	header.SetMode(0755)

	entry, err := writer.CreateHeader(header)
	if err != nil {
		_ = file.Close()
		t.Fatalf("create zip entry: %v", err)
	}

	if _, err = entry.Write([]byte("chrome binary")); err != nil {
		_ = file.Close()
		t.Fatalf("write zip entry: %v", err)
	}

	if err = writer.Close(); err != nil {
		_ = file.Close()
		t.Fatalf("close zip writer: %v", err)
	}

	if err = file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	destination := filepath.Join(t.TempDir(), "chrome", "120.0.6099.109")

	installer := New()

	if err = installer.Install(archive, destination); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	chromePath := filepath.Join(destination, "chrome-linux64", "chrome")

	info, err := os.Stat(chromePath)
	if err != nil {
		t.Fatalf("stat installed file: %v", err)
	}

	permissions := info.Mode().Perm()

	if permissions != 0755 {
		t.Fatalf("expected permissions 0755, got %o", permissions)
	}
}
