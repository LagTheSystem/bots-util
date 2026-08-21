package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"bots-util/internal/config"
	"bots-util/internal/pkg/download"
	"bots-util/internal/pkg/syscmd"
)

type mockDownloader struct {
	downloadFunc func(ctx context.Context, url, dest string, cb func(int64, int64)) error
}

func (m *mockDownloader) Download(ctx context.Context, url, dest string, cb download.ProgressCallback) error {
	return m.downloadFunc(ctx, url, dest, cb)
}

type mockCommandRunner struct {
	runFunc func(ctx context.Context, cmd syscmd.Command, cb func(string)) error
}

func (m *mockCommandRunner) Run(ctx context.Context, cmd syscmd.Command, cb syscmd.LineCallback) error {
	return m.runFunc(ctx, cmd, cb)
}

func TestDesktopServiceRun(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "target")
	os.MkdirAll(destDir, 0755)

	downloader := &mockDownloader{
		downloadFunc: func(ctx context.Context, url, dest string, cb func(int64, int64)) error {
			cb(100, 100)
			return os.WriteFile(dest, []byte("fake zip"), 0644)
		},
	}
	runner := &mockCommandRunner{
		runFunc: func(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
			cb("extracting...")
			return nil
		},
	}

	cfg := config.DesktopConfig{
		TargetFolder: destDir,
		ZipURL:       "https://example.com/camp.zip",
		TempDir:      t.TempDir(),
	}

	svc := NewDesktopService(cfg, downloader, runner)
	var msgs []string
	err := svc.Run(context.Background(), func(msg string, pct float64) {
		msgs = append(msgs, msg)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected progress messages")
	}
}

func TestDesktopServiceDownloadError(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "target")
	os.MkdirAll(destDir, 0755)

	downloader := &mockDownloader{
		downloadFunc: func(ctx context.Context, url, dest string, cb func(int64, int64)) error {
			return errors.New("network error")
		},
	}
	runner := &mockCommandRunner{
		runFunc: func(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
			return nil
		},
	}

	cfg := config.DesktopConfig{
		TargetFolder: destDir,
		ZipURL:       "https://example.com/camp.zip",
		TempDir:      t.TempDir(),
	}

	svc := NewDesktopService(cfg, downloader, runner)
	err := svc.Run(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for download failure, got nil")
	}
}
