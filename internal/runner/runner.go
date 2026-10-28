package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, executable string, args []string) error {
	cmd := exec.CommandContext(ctx, executable, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start browser: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("browser process: %w", err)
	}

	return nil
}
