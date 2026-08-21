package syscmd

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type LineCallback func(line string)

type Command struct {
	Name    string
	Exe     string
	Args    []string
	Timeout time.Duration
}

type CommandRunner interface {
	Run(ctx context.Context, cmd Command, cb LineCallback) error
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, cmd Command, cb LineCallback) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	c := exec.CommandContext(timeoutCtx, cmd.Exe, cmd.Args...)

	stdout, err := c.StdoutPipe()
	if err != nil {
		return fmt.Errorf("syscmd: %s: stdout pipe: %w", cmd.Name, err)
	}
	c.Stderr = c.Stdout

	if err := c.Start(); err != nil {
		return fmt.Errorf("syscmd: %s: start: %w", cmd.Name, err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if cb != nil {
			cb(scanner.Text())
		}
	}

	waitErr := c.Wait()

	if timeoutCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("syscmd: %s: timed out after %s", cmd.Name, cmd.Timeout)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("syscmd: %s: cancelled", cmd.Name)
	}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			return fmt.Errorf("syscmd: %s: exit code %d: %w", cmd.Name, exitErr.ExitCode(), waitErr)
		}
		return fmt.Errorf("syscmd: %s: %w", cmd.Name, waitErr)
	}

	return nil
}