package downloader

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloader_Download(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.Handler
		cancelContext bool
		wantContent   string
		wantErr       bool
	}{
		{
			name: "success",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("chrome archive content"))
			}),
			wantContent: "chrome archive content",
		},
		{
			name: "http error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			}),
			wantErr: true,
		},
		{
			name:          "context canceled",
			cancelContext: true,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			destination := filepath.Join(t.TempDir(), "chrome.zip")

			ctx := context.Background()

			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			downloader := New(server.Client())

			err := downloader.Download(ctx, server.URL, destination, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, err := os.ReadFile(destination)
			if err != nil {
				t.Fatalf("read downloaded file: %v", err)
			}

			if string(content) != tt.wantContent {
				t.Fatalf("expected content %q, got %q", tt.wantContent, string(content))
			}
		})
	}
}

func TestDownloader_CancelDuringDownload(t *testing.T) {
	firstChunkWritten := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("partial archive"))
		w.(http.Flusher).Flush()

		close(firstChunkWritten)
		<-r.Context().Done()
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "chrome.zip")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	downloader := New(server.Client())

	result := make(chan error, 1)

	go func() {
		result <- downloader.Download(ctx, server.URL, destination, nil)
	}()

	select {
	case <-firstChunkWritten:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not start sending data")
	}

	cancel()

	select {
	case err := <-result:
		if err == nil {
			t.Fatal("Download() error = nil, want cancellation error")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Download() error = %v, want error wrapping %v", err, context.Canceled)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("Download() did not return after cancellation")
	}

	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("destination should not exist after cancellation, stat error = %v", err)
	}

	tempFiles, err := filepath.Glob(filepath.Join(filepath.Dir(destination), ".bmv-download-*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(tempFiles) != 0 {
		t.Errorf("temporary files remain after cancellation: %v", tempFiles)
	}
}
