package service

import (
	"context"
	"errors"
	"testing"

	"bots-util/internal/config"
	"bots-util/internal/pkg/syscmd"
)

func TestRepairServiceRunAll(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
			cb(cmd.Name + " output line 1")
			cb(cmd.Name + " output line 2")
			return nil
		},
	}

	cfg := config.RepairConfig{
		Commands: []config.RepairCommand{
			{Name: "DISM", Exe: "dism", Args: []string{"/Online", "/Cleanup-Image", "/RestoreHealth"}},
			{Name: "SFC", Exe: "sfc", Args: []string{"/scannow"}},
		},
	}

	svc := NewRepairService(cfg, runner)
	var msgs []string
	err := svc.RunAll(context.Background(), func(msg string, pct float64) {
		msgs = append(msgs, msg)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(msgs) < 4 {
		t.Fatalf("expected at least 4 progress messages (2 start + 2 output), got %d", len(msgs))
	}
}

func TestRepairServiceRunAllFailSoft(t *testing.T) {
	callCount := 0
	runner := &mockCommandRunner{
		runFunc: func(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
			callCount++
			cb(cmd.Name)
			if callCount == 1 {
				return errors.New("DISM failed")
			}
			return nil
		},
	}

	cfg := config.RepairConfig{
		Commands: []config.RepairCommand{
			{Name: "DISM", Exe: "dism", Args: []string{"/Online", "/Cleanup-Image", "/RestoreHealth"}},
			{Name: "SFC", Exe: "sfc", Args: []string{"/scannow"}},
		},
	}

	svc := NewRepairService(cfg, runner)
	err := svc.RunAll(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error when DISM fails, got nil")
	}
	if callCount != 2 {
		t.Fatalf("expected both commands to run (fail-soft), got %d calls", callCount)
	}
}

func TestRepairServiceCancellation(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
			return nil
		},
	}

	cfg := config.RepairConfig{
		Commands: []config.RepairCommand{
			{Name: "DISM", Exe: "dism", Args: []string{"/Online", "/Cleanup-Image", "/RestoreHealth"}},
			{Name: "SFC", Exe: "sfc", Args: []string{"/scannow"}},
		},
	}

	svc := NewRepairService(cfg, runner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.RunAll(ctx, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
