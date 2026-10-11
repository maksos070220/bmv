package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ProgressFunc func(downloaded, total int64)

type Downloader struct {
	client *http.Client
}

func New(client *http.Client) *Downloader {
	if client == nil {
		client = http.DefaultClient
	}

	return &Downloader{
		client: client,
	}
}

func (d *Downloader) Download(ctx context.Context, url string, destination string, onProgress ProgressFunc) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	if err = os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(destination), ".bmv-download-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	tempPath := tempFile.Name()

	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	var downloaded int64
	total := resp.ContentLength

	buffer := make([]byte, 32*1024)

	for {
		if err = ctx.Err(); err != nil {
			return err
		}

		n, readErr := resp.Body.Read(buffer)

		if n > 0 {
			written, writeErr := tempFile.Write(buffer[:n])
			downloaded += int64(written)

			if writeErr != nil {
				return fmt.Errorf("write download: %w", writeErr)
			}

			if written != n {
				return io.ErrShortWrite
			}

			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return fmt.Errorf("read download: %w", readErr)
		}
	}

	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}

	if err = os.Rename(tempPath, destination); err != nil {
		return fmt.Errorf("move downloaded file: %w", err)
	}

	return nil
}
