package syscmd

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunnerRunSuccess(t *testing.T) {
	r := NewRunner()
	cmd := Command{
		Name:    "echo",
		Exe:     "echo",
		Args:    []string{"hello"},
		Timeout: 5 * time.Second,
	}

	var lines []string
	err := r.Run(context.Background(), cmd, func(line string) {
		lines = append(lines, line)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected at least one output line")
	}
	if !strings.Contains(strings.Join(lines, "\n"), "hello") {
		t.Fatalf("expected output to contain 'hello', got %v", lines)
	}
}

func TestRunnerRunFailure(t *testing.T) {
	r := NewRunner()
	cmd := Command{
		Name:    "false",
		Exe:     "false",
		Args:    []string{},
		Timeout: 5 * time.Second,
	}

	err := r.Run(context.Background(), cmd, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}

func TestRunnerRunCancellation(t *testing.T) {
	r := NewRunner()
	cmd := Command{
		Name:    "sleep",
		Exe:     "sleep",
		Args:    []string{"30"},
		Timeout: 5 * time.Minute,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Run(ctx, cmd, nil)
	}()
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error from cancellation")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cancellation")
	}
}

func TestRunnerRunTimeout(t *testing.T) {
	r := NewRunner()
	cmd := Command{
		Name:    "sleep",
		Exe:     "sleep",
		Args:    []string{"30"},
		Timeout: 100 * time.Millisecond,
	}

	err := r.Run(context.Background(), cmd, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
