package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bots-util/internal/config"
	"bots-util/internal/pkg/syscmd"
)

type RepairService struct {
	config config.RepairConfig
	runner syscmd.CommandRunner
}

func NewRepairService(cfg config.RepairConfig, runner syscmd.CommandRunner) *RepairService {
	return &RepairService{config: cfg, runner: runner}
}

func (s *RepairService) RunAll(ctx context.Context, progress func(string, float64)) error {
	total := len(s.config.Commands)
	if total == 0 {
		return nil
	}

	var errs []string

	for i, cmd := range s.config.Commands {
		select {
		case <-ctx.Done():
			return fmt.Errorf("repair: cancelled")
		default:
		}

		if progress != nil {
			progress(fmt.Sprintf("running %s...", cmd.Name), float64(i)/float64(total))
		}

		timeout := 30 * time.Minute
		if cmd.Timeout != "" {
			if d, err := time.ParseDuration(cmd.Timeout); err == nil {
				timeout = d
			}
		}

		sc := syscmd.Command{
			Name:    cmd.Name,
			Exe:     cmd.Exe,
			Args:    cmd.Args,
			Timeout: timeout,
		}

		err := s.runner.Run(ctx, sc, func(line string) {
			if progress != nil {
				progress(fmt.Sprintf("[%s] %s", cmd.Name, line), float64(i)/float64(total))
			}
		})

		if err != nil {
			msg := fmt.Sprintf("%s failed: %v", cmd.Name, err)
			errs = append(errs, msg)
			if progress != nil {
				progress(msg, float64(i+1)/float64(total))
			}
		} else {
			if progress != nil {
				progress(fmt.Sprintf("%s completed successfully", cmd.Name), float64(i+1)/float64(total))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("repair: %d command(s) failed: %s", len(errs), strings.Join(errs, "; "))
	}

	if progress != nil {
		progress("all repair commands completed", 1.0)
	}
	return nil
}
