package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestClientDownload(t *testing.T) {
	content := []byte("hello download world")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.txt")
	client := NewClient()

	var lastBytes, lastTotal int64
	err := client.Download(context.Background(), ts.URL, dest, func(bytes, total int64) {
		lastBytes = bytes
		lastTotal = total
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("expected %q, got %q", content, data)
	}
	if lastBytes != int64(len(content)) {
		t.Fatalf("expected lastBytes %d, got %d", len(content), lastBytes)
	}
	if lastTotal != int64(len(content)) {
		t.Fatalf("expected lastTotal %d, got %d", len(content), lastTotal)
	}
}

func TestClientDownloadServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.txt")
	client := NewClient()

	err := client.Download(context.Background(), ts.URL, dest, nil)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestClientDownloadCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.txt")
	client := NewClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Download(ctx, ts.URL, dest, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestClientDownloadCreatesParentDirs(t *testing.T) {
	content := []byte("test")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "sub", "nested", "file.txt")
	client := NewClient()

	err := client.Download(context.Background(), ts.URL, dest, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("expected %q, got %q", content, data)
	}
}
