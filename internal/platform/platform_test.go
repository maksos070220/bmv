package platform

import (
	"runtime"
	"testing"
)

func TestCurrent(t *testing.T) {
	got, err := Current()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.OS != runtime.GOOS {
		t.Errorf("expected OS %q, got %q", runtime.GOOS, got.OS)
	}

	if got.Architecture != runtime.GOARCH {
		t.Errorf("expected architecture %q, got %q", runtime.GOARCH, got.Architecture)
	}
}
