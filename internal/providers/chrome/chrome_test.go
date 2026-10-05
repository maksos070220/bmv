package chrome

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maksos070220/bmv/internal/browser"
)

func TestProvider_Resolve(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		platform      browser.Platform
		expectedVer   string
		expectedURL   string
		expectedError string
	}{
		{
			name:    "linux amd64",
			version: "120",
			platform: browser.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expectedVer: "120.0.6099.109",
			expectedURL: "https://example.com/chrome-linux64.zip",
		},
		{
			name:    "windows amd64",
			version: "120",
			platform: browser.Platform{
				OS:           "windows",
				Architecture: "amd64",
			},
			expectedVer: "120.0.6099.109",
			expectedURL: "https://example.com/chrome-win64.zip",
		},
		{
			name:    "windows 386",
			version: "120",
			platform: browser.Platform{
				OS:           "windows",
				Architecture: "386",
			},
			expectedVer: "120.0.6099.109",
			expectedURL: "https://example.com/chrome-win32.zip",
		},
		{
			name:    "unsupported version",
			version: "999",
			platform: browser.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			expectedError: `Chrome milestone "999" not found`,
		},
		{
			name:    "unsupported platform",
			version: "120",
			platform: browser.Platform{
				OS:           "linux",
				Architecture: "386",
			},
			expectedError: "unsupported platform: linux/386",
		},
	}

	t.Run("resolve", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				manifest := map[string]any{
					"milestones": map[string]any{
						"120": map[string]any{
							"version":  "120.0.6099.109",
							"revision": "123456",
							"downloads": map[string]any{
								"chrome": []map[string]string{
									{
										"platform": "linux64",
										"url":      "https://example.com/chrome-linux64.zip",
									},
									{
										"platform": "win64",
										"url":      "https://example.com/chrome-win64.zip",
									},
									{
										"platform": "win32",
										"url":      "https://example.com/chrome-win32.zip",
									},
								},
							},
						},
					},
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(manifest)
			}),
		)
		defer server.Close()

		provider := NewProvider(server.Client())
		provider.latestVersionsURL = server.URL

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				release, err := provider.Resolve(context.Background(), tt.version, tt.platform)

				if tt.expectedError != "" {
					if err == nil {
						t.Fatalf("expected error %q, got nil", tt.expectedError)
					}

					if err.Error() != tt.expectedError {
						t.Fatalf("expected error %q, got %q", tt.expectedError, err.Error())
					}

					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if release.Version != tt.expectedVer {
					t.Errorf("expected version %q, got %q", tt.expectedVer, release.Version)
				}

				if release.DownloadURL != tt.expectedURL {
					t.Errorf("expected URL %q, got %q", tt.expectedURL, release.DownloadURL)
				}

				if release.Browser != browser.Chrome {
					t.Errorf("expected browser %q, got %q", browser.Chrome, release.Browser)
				}
			})
		}
	})
}
