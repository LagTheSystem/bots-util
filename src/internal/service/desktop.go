package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bots-util/internal/config"
	"bots-util/internal/pkg/download"
	"bots-util/internal/pkg/syscmd"
)

type DesktopService struct {
	config     config.DesktopConfig
	downloader download.Downloader
	runner     syscmd.CommandRunner
}

func NewDesktopService(cfg config.DesktopConfig, dl download.Downloader, sc syscmd.CommandRunner) *DesktopService {
	return &DesktopService{config: cfg, downloader: dl, runner: sc}
}

func (s *DesktopService) Run(ctx context.Context, progress func(string, float64)) error {
	if progress != nil {
		progress("clearing target folder...", 0.0)
	}

	if err := s.clearFolder(ctx); err != nil {
		return fmt.Errorf("desktop: clear folder: %w", err)
	}

	zipPath := filepath.Join(s.config.TempDir, "camp.zip")
	if progress != nil {
		progress("downloading...", 0.05)
	}

	if err := s.downloader.Download(ctx, s.config.ZipURL, zipPath, func(bytes, total int64) {
		if progress != nil {
			if total > 0 {
				pct := 0.05 + 0.65*(float64(bytes)/float64(total))
				progress(fmt.Sprintf("downloading... %d/%d bytes", bytes, total), pct)
			} else {
				progress(fmt.Sprintf("downloading... %d bytes", bytes), 0.05)
			}
		}
	}); err != nil {
		return fmt.Errorf("desktop: download zip: %w", err)
	}

	defer os.Remove(zipPath)

	if progress != nil {
		progress("extracting...", 0.70)
	}

	extractCmd := syscmd.Command{
		Name:    "extract",
		Exe:     "powershell",
		Args:    []string{"-Command", fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipPath, s.config.TargetFolder)},
		Timeout: 5 * time.Minute,
	}

	if err := s.runner.Run(ctx, extractCmd, func(line string) {
		if progress != nil {
			progress("extracting: "+line, 0.70)
		}
	}); err != nil {
		return fmt.Errorf("desktop: extract zip: %w", err)
	}

	if progress != nil {
		progress("done", 1.0)
	}

	return nil
}

func (s *DesktopService) clearFolder(ctx context.Context) error {
	entries, err := os.ReadDir(s.config.TargetFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(s.config.TargetFolder, 0755)
		}
		return err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		p := filepath.Join(s.config.TargetFolder, entry.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}

	return nil
}
