package runner

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestRunnerRunSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test requires a Unix shell")
	}

	outputFile := filepath.Join(t.TempDir(), "args.txt")

	args := []string{
		"-c",
		`printf '%s\n' "$@" > "$OUTPUT_FILE"`,
		"test-process",
		"first",
		"second argument",
		"--headless",
	}

	t.Setenv("OUTPUT_FILE", outputFile)

	r := New()

	err := r.Run(context.Background(), "/bin/sh", args)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	gotBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	got := splitLines(string(gotBytes))
	want := []string{
		"first",
		"second argument",
		"--headless",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("process arguments = %q, want %q", got, want)
	}
}

func TestRunnerRunCommandNotFound(t *testing.T) {
	r := New()

	err := r.Run(
		context.Background(),
		filepath.Join(t.TempDir(), "does-not-exist"),
		nil,
	)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestRunnerRunCanceledContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test requires a Unix shell")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New()

	err := r.Run(ctx, "/bin/sh", []string{
		"-c",
		"sleep 10",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestRunnerRunContextCanceledDuringExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test requires a Unix shell")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r := New()

	err := r.Run(ctx, "/bin/sh", []string{
		"-c",
		"sleep 10",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Fatalf("context error = %v, want %v", ctx.Err(), context.DeadlineExceeded)
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	var lines []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}

	return lines
}
