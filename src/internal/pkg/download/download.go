package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ProgressCallback func(bytesDownloaded, totalBytes int64)

type Downloader interface {
	Download(ctx context.Context, url, dest string, cb ProgressCallback) error
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{}}
}

func (c *Client) Download(ctx context.Context, url, dest string, cb ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("download: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s: HTTP %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("download: create dirs: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("download: create file %s: %w", dest, err)
	}
	defer f.Close()

	totalBytes := resp.ContentLength
	var written int64

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			os.Remove(dest)
			return fmt.Errorf("download: %s: cancelled", url)
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, writeErr := f.Write(buf[:n])
			if writeErr != nil {
				os.Remove(dest)
				return fmt.Errorf("download: write to %s: %w", dest, writeErr)
			}
			written += int64(wn)
			if cb != nil {
				cb(written, totalBytes)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			os.Remove(dest)
			return fmt.Errorf("download: read body: %w", readErr)
		}
	}

	return nil
}
