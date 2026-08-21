# bots-util Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Windows desktop utility with four independent actions (desktop folder restore, Chrome extension install, system repair, policy/registry settings) using Go + Wails v2 + Svelte.

**Architecture:** Three-layer Go backend: shared infrastructure packages (`pkg/syscmd`, `pkg/registry`, `pkg/download`) → service layer (`service/desktop`, `service/chrome`, `service/repair`, `service/policy`) → thin handler bound to Wails frontend. Services accept interfaces for testability; a Svelte dashboard renders four independent action cards with live progress.

**Tech Stack:** Go 1.26.5, Wails v2, Svelte + Vite, `golang.org/x/sys/windows/registry`

**Spec:** `docs/superpowers/specs/2025-08-21-bots-util-design.md`

## Global Constraints

- Module name: `bots-util`, root directory: `src/`
- All configuration is hardcoded in `internal/config/config.go` — no external config files
- Windows-only — registry and system commands are Windows-native
- Services must accept interfaces (not concrete types) for testability
- Tests use standard library `testing`; mocks are inline structs, no external mocking library
- Wails v2 stable (`github.com/wailsapp/wails/v2`)
- Frontend: Svelte with Vite, Wails-generated JS bindings (`window.go.main.App.*`)

---

### Task 1: Project Scaffolding

**Files:**
- Create: `src/wails.json`
- Create: `src/frontend/index.html`
- Create: `src/frontend/package.json`
- Create: `src/frontend/vite.config.js`
- Create: `src/frontend/src/main.js`
- Create: `src/frontend/src/app.css`
- Modify: `src/go.mod` (add Wails dependency)
- Delete: `src/internal/git.go`
- Create empty dirs: `src/internal/config/`, `src/internal/handler/`, `src/internal/service/`, `src/internal/pkg/registry/`, `src/internal/pkg/syscmd/`, `src/internal/pkg/download/`, `src/frontend/src/components/`

**Interfaces:**
- Consumes: nothing
- Produces: project structure, `go.mod` with Wails dependency, `wails.json` config

- [ ] **Step 1: Create wails.json**

```json
{
  "name": "bots-util",
  "outputfilename": "bots-util",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "",
    "email": ""
  }
}
```

- [ ] **Step 2: Create frontend/package.json**

```json
{
  "name": "frontend",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^3.0.0",
    "svelte": "^4.2.0",
    "vite": "^5.0.0"
  }
}
```

- [ ] **Step 3: Create frontend/vite.config.js**

```js
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()]
})
```

- [ ] **Step 4: Create frontend/index.html**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>bots-util</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

- [ ] **Step 5: Create frontend/src/main.js**

```js
import App from './App.svelte'
import './app.css'

const app = new App({
  target: document.getElementById('app')
})

export default app
```

- [ ] **Step 6: Create frontend/src/app.css**

```css
*, *::before, *::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #1a1a2e;
  color: #e0e0e0;
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  padding: 24px;
}
```

- [ ] **Step 7: Update go.mod with Wails dependency**

Run: `cd src && go get github.com/wailsapp/wails/v2@latest`

- [ ] **Step 8: Remove old internal/git.go**

Run: `rm src/internal/git.go`

- [ ] **Step 9: Create required directories**

Run: `mkdir -p src/internal/config src/internal/handler src/internal/service src/internal/pkg/registry src/internal/pkg/syscmd src/internal/pkg/download src/frontend/src/components`

- [ ] **Step 10: Commit**

```bash
git add src/wails.json src/frontend/ src/go.mod src/go.sum
git rm src/internal/git.go
git commit -m "feat: scaffold Wails v2 project with Svelte frontend"
```

---

### Task 2: Config Package

**Files:**
- Create: `src/internal/config/config.go`

**Interfaces:**
- Consumes: nothing
- Produces: `DesktopConfig`, `ChromeConfig`, `RepairConfig`, `RepairCommand`, `PolicyConfig`, `RegistryEntry` types; `DefaultDesktopConfig()`, `DefaultChromeConfig()`, `DefaultRepairConfig()`, `DefaultPolicyConfig()` functions

- [ ] **Step 1: Write config.go**

```go
package config

type DesktopConfig struct {
	TargetFolder string
	ZipURL       string
	TempDir      string
}

type ChromeConfig struct {
	ExtensionID  string
	UpdateURL    string
	ForceInstall bool
}

type RepairConfig struct {
	Commands []RepairCommand
}

type RepairCommand struct {
	Name    string
	Exe     string
	Args    []string
	Timeout string // e.g. "30m"
}

type PolicyConfig struct {
	RegistryEntries []RegistryEntry
}

type RegistryEntry struct {
	KeyPath string
	Name    string
	Value   any
	Type    uint32 // registry.DWORD, registry.SZ, etc.
}

func DefaultDesktopConfig() DesktopConfig {
	return DesktopConfig{
		TargetFolder: `C:\Users\Public\Desktop\CampShortcuts`,
		ZipURL:       "",
		TempDir:      `C:\Windows\Temp\bots-util`,
	}
}

func DefaultChromeConfig() ChromeConfig {
	return ChromeConfig{
		ExtensionID:  "",
		UpdateURL:    "https://clients2.google.com/service/update2/crx",
		ForceInstall: true,
	}
}

func DefaultRepairConfig() RepairConfig {
	return RepairConfig{
		Commands: []RepairCommand{
			{
				Name:    "DISM RestoreHealth",
				Exe:     "dism",
				Args:    []string{"/Online", "/Cleanup-Image", "/RestoreHealth"},
				Timeout: "30m",
			},
			{
				Name:    "SFC ScanNow",
				Exe:     "sfc",
				Args:    []string{"/scannow"},
				Timeout: "15m",
			},
		},
	}
}

func DefaultPolicyConfig() PolicyConfig {
	return PolicyConfig{
		RegistryEntries: []RegistryEntry{},
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd src && go build ./internal/config/`

- [ ] **Step 3: Commit**

```bash
git add src/internal/config/config.go
git commit -m "feat: add config package with hardcoded defaults"
```

---

### Task 3: syscmd Package

**Files:**
- Create: `src/internal/pkg/syscmd/syscmd.go`
- Create: `src/internal/pkg/syscmd/syscmd_test.go`

**Interfaces:**
- Consumes: `config.RepairCommand` (by similarity — this package defines its own `Command` type)
- Produces: `Command` struct, `LineCallback` type, `CommandRunner` interface, `Runner` struct, `NewRunner()`, `Runner.Run()`

- [ ] **Step 1: Write the failing test**

```go
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
		Exe:     "cmd",
		Args:    []string{"/c", "echo hello"},
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
		Name:    "exit",
		Exe:     "cmd",
		Args:    []string{"/c", "exit 1"},
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
		Exe:     "cmd",
		Args:    []string{"/c", "timeout /t 30"},
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
		Exe:     "cmd",
		Args:    []string{"/c", "timeout /t 30"},
		Timeout: 100 * time.Millisecond,
	}

	err := r.Run(context.Background(), cmd, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/pkg/syscmd/ -v`
Expected: compilation error — `CommandRunner`/`Runner`/`Run` not defined

- [ ] **Step 3: Write syscmd.go**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/pkg/syscmd/ -v`
Expected: all 4 tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/pkg/syscmd/
git commit -m "feat: add syscmd package with command runner and streaming output"
```

---

### Task 4: registry Package

**Files:**
- Create: `src/internal/pkg/registry/registry.go`
- Create: `src/internal/pkg/registry/registry_test.go`

**Interfaces:**
- Consumes: `golang.org/x/sys/windows/registry`
- Produces: `Registry` interface, `RegistryImpl` struct, `New()`, all registry methods

- [ ] **Step 1: Write the failing test**

```go
package registry

import (
	"testing"
)

func TestRegistryInterface(t *testing.T) {
	var r Registry = New()
	_ = r
}

func TestSetAndDeleteDWORD(t *testing.T) {
	r := New()
	testKey := `SOFTWARE\bots-util-test`
	testName := "TestValue"
	var testValue uint32 = 42

	err := r.SetDWORD(testKey, testName, testValue)
	if err != nil {
		t.Fatalf("SetDWORD failed: %v", err)
	}

	exists, err := r.ValueExists(testKey, testName)
	if err != nil {
		t.Fatalf("ValueExists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected value to exist after SetDWORD")
	}

	err = r.Delete(testKey, testName)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, err = r.ValueExists(testKey, testName)
	if err != nil {
		t.Fatalf("ValueExists after delete failed: %v", err)
	}
	if exists {
		t.Fatal("expected value to not exist after Delete")
	}
}

func TestSetAndDeleteString(t *testing.T) {
	r := New()
	testKey := `SOFTWARE\bots-util-test`
	testName := "TestString"
	testValue := "hello world"

	err := r.SetString(testKey, testName, testValue)
	if err != nil {
		t.Fatalf("SetString failed: %v", err)
	}

	exists, err := r.ValueExists(testKey, testName)
	if err != nil {
		t.Fatalf("ValueExists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected value to exist after SetString")
	}

	err = r.Delete(testKey, testName)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestKeyExists(t *testing.T) {
	r := New()
	exists, err := r.KeyExists(`SOFTWARE\Microsoft\Windows`)
	if err != nil {
		t.Fatalf("KeyExists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected SOFTWARE\\Microsoft\\Windows to exist")
	}

	exists, err = r.KeyExists(`SOFTWARE\definitely-does-not-exist-xyz123`)
	if err != nil {
		t.Fatalf("KeyExists for non-existent key: %v", err)
	}
	if exists {
		t.Fatal("expected non-existent key to not exist")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/pkg/registry/ -v`
Expected: compilation error — `Registry` interface not defined

- [ ] **Step 3: Write registry.go**

```go
package registry

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	DWORD = registry.DWORD
	SZ    = registry.SZ
)

type Registry interface {
	SetDWORD(key, name string, value uint32) error
	SetString(key, name, value string) error
	Delete(key, name string) error
	KeyExists(key string) (bool, error)
	ValueExists(key, name string) (bool, error)
}

type RegistryImpl struct{}

func New() *RegistryImpl {
	return &RegistryImpl{}
}

func (r *RegistryImpl) SetDWORD(key, name string, value uint32) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("registry: open key %s: %w", key, err)
	}
	defer k.Close()

	if err := k.SetDWordValue(name, value); err != nil {
		return fmt.Errorf("registry: set DWORD %s\\%s: %w", key, name, err)
	}
	return nil
}

func (r *RegistryImpl) SetString(key, name, value string) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("registry: open key %s: %w", key, err)
	}
	defer k.Close()

	if err := k.SetStringValue(name, value); err != nil {
		return fmt.Errorf("registry: set string %s\\%s: %w", key, name, err)
	}
	return nil
}

func (r *RegistryImpl) Delete(key, name string) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("registry: open key for delete %s: %w", key, err)
	}
	defer k.Close()

	if err := k.DeleteValue(name); err != nil {
		return fmt.Errorf("registry: delete value %s\\%s: %w", key, name, err)
	}
	return nil
}

func (r *RegistryImpl) KeyExists(key string) (bool, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE)
	if err != nil {
		return false, nil
	}
	k.Close()
	return true, nil
}

func (r *RegistryImpl) ValueExists(key, name string) (bool, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE)
	if err != nil {
		return false, nil
	}
	defer k.Close()

	_, _, err = k.GetValue(name, nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}
```

- [ ] **Step 4: Add golang.org/x/sys dependency**

Run: `cd src && go get golang.org/x/sys/windows/registry`

- [ ] **Step 5: Run tests to verify they pass** (Windows only — skip on macOS)

Run: `cd src && go test ./internal/pkg/registry/ -v -run TestRegistryInterface`
Expected: builds and compiles (full integration tests require Windows)

- [ ] **Step 6: Commit**

```bash
git add src/internal/pkg/registry/ src/go.mod src/go.sum
git commit -m "feat: add registry package wrapping Windows registry API"
```

---

### Task 5: download Package

**Files:**
- Create: `src/internal/pkg/download/download.go`
- Create: `src/internal/pkg/download/download_test.go`

**Interfaces:**
- Consumes: `net/http`, `net/http/httptest` (test only)
- Produces: `ProgressCallback` type, `Downloader` interface, `Client` struct, `NewClient()`, `Client.Download()`

- [ ] **Step 1: Write the failing test**

```go
package download

import (
	"context"
	"fmt"
	"io"
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/pkg/download/ -v`
Expected: compilation error — `Downloader`/`Client` not defined

- [ ] **Step 3: Write download.go**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/pkg/download/ -v`
Expected: all 4 tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/pkg/download/
git commit -m "feat: add download package with progress callback"
```

---

### Task 6: Policy Service

**Files:**
- Create: `src/internal/service/policy.go`
- Create: `src/internal/service/policy_test.go`

**Interfaces:**
- Consumes: `config.PolicyConfig`, `config.RegistryEntry`, `registry.Registry` (interface)
- Produces: `PolicyService` struct, `NewPolicyService()`, `PolicyService.Apply()`

- [ ] **Step 1: Write the failing test**

```go
package service

import (
	"context"
	"errors"
	"testing"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

type mockRegistry struct {
	setDWORDErr  error
	setStringErr error
	dwordCalls   []struct {
		key   string
		name  string
		value uint32
	}
	stringCalls []struct {
		key   string
		name  string
		value string
	}
}

func (m *mockRegistry) SetDWORD(key, name string, value uint32) error {
	m.dwordCalls = append(m.dwordCalls, struct {
		key   string
		name  string
		value uint32
	}{key, name, value})
	return m.setDWORDErr
}

func (m *mockRegistry) SetString(key, name, value string) error {
	m.stringCalls = append(m.stringCalls, struct {
		key   string
		name  string
		value string
	}{key, name, value})
	return m.setStringErr
}

func (m *mockRegistry) Delete(key, name string) error {
	return nil
}

func (m *mockRegistry) KeyExists(key string) (bool, error) {
	return false, nil
}

func (m *mockRegistry) ValueExists(key, name string) (bool, error) {
	return false, nil
}

func TestPolicyServiceApply(t *testing.T) {
	mock := &mockRegistry{}
	cfg := config.PolicyConfig{
		RegistryEntries: []config.RegistryEntry{
			{
				KeyPath: `SOFTWARE\Policies\Test`,
				Name:    "Setting1",
				Value:   uint32(1),
				Type:    registry.DWORD,
			},
			{
				KeyPath: `SOFTWARE\Policies\Test`,
				Name:    "Setting2",
				Value:   "enabled",
				Type:    registry.SZ,
			},
		},
	}

	svc := NewPolicyService(cfg, mock)
	progress := make([]string, 0)
	err := svc.Apply(context.Background(), func(msg string, idx int) {
		progress = append(progress, msg)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mock.dwordCalls) != 1 {
		t.Fatalf("expected 1 DWORD call, got %d", len(mock.dwordCalls))
	}
	if mock.dwordCalls[0].name != "Setting1" {
		t.Fatalf("expected DWORD name Setting1, got %s", mock.dwordCalls[0].name)
	}
	if mock.dwordCalls[0].value != uint32(1) {
		t.Fatalf("expected DWORD value 1, got %d", mock.dwordCalls[0].value)
	}
	if len(mock.stringCalls) != 1 {
		t.Fatalf("expected 1 string call, got %d", len(mock.stringCalls))
	}
	if mock.stringCalls[0].name != "Setting2" {
		t.Fatalf("expected string name Setting2, got %s", mock.stringCalls[0].name)
	}
	if mock.stringCalls[0].value != "enabled" {
		t.Fatalf("expected string value enabled, got %s", mock.stringCalls[0].value)
	}
	if len(progress) != 2 {
		t.Fatalf("expected 2 progress calls, got %d", len(progress))
	}
}

func TestPolicyServiceApplyEmptyEntries(t *testing.T) {
	mock := &mockRegistry{}
	cfg := config.PolicyConfig{RegistryEntries: []config.RegistryEntry{}}

	svc := NewPolicyService(cfg, mock)
	err := svc.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error for empty entries, got %v", err)
	}
}

func TestPolicyServiceApplyError(t *testing.T) {
	mock := &mockRegistry{setDWORDErr: errors.New("registry access denied")}
	cfg := config.PolicyConfig{
		RegistryEntries: []config.RegistryEntry{
			{
				KeyPath: `SOFTWARE\Policies\Test`,
				Name:    "Setting1",
				Value:   uint32(1),
				Type:    registry.DWORD,
			},
		},
	}

	svc := NewPolicyService(cfg, mock)
	err := svc.Apply(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/service/ -v -run TestPolicy`
Expected: compilation error — `PolicyService` not defined

- [ ] **Step 3: Write policy.go**

```go
package service

import (
	"context"
	"fmt"
	"strings"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

type PolicyService struct {
	config config.PolicyConfig
	reg    registry.Registry
}

func NewPolicyService(cfg config.PolicyConfig, reg registry.Registry) *PolicyService {
	return &PolicyService{config: cfg, reg: reg}
}

func (s *PolicyService) Apply(ctx context.Context, progress func(string, int)) error {
	if len(s.config.RegistryEntries) == 0 {
		if progress != nil {
			progress("no registry entries to apply", 0)
		}
		return nil
	}

	var errs []string

	for i, entry := range s.config.RegistryEntries {
		var err error
		switch entry.Type {
		case registry.DWORD:
			val, ok := entry.Value.(uint32)
			if !ok {
				errs = append(errs, fmt.Sprintf("policy: %s: invalid DWORD value type", entry.Name))
				continue
			}
			err = s.reg.SetDWORD(entry.KeyPath, entry.Name, val)
		case registry.SZ:
			val, ok := entry.Value.(string)
			if !ok {
				errs = append(errs, fmt.Sprintf("policy: %s: invalid string value type", entry.Name))
				continue
			}
			err = s.reg.SetString(entry.KeyPath, entry.Name, val)
		default:
			errs = append(errs, fmt.Sprintf("policy: %s: unsupported registry type %d", entry.Name, entry.Type))
			continue
		}

		if err != nil {
			msg := fmt.Sprintf("failed to set %s\\%s: %v", entry.KeyPath, entry.Name, err)
			errs = append(errs, msg)
			if progress != nil {
				progress(msg, i)
			}
		} else {
			msg := fmt.Sprintf("set %s\\%s", entry.KeyPath, entry.Name)
			if progress != nil {
				progress(msg, i)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("policy: %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/service/ -v -run TestPolicy`
Expected: all 3 tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/service/policy.go src/internal/service/policy_test.go
git commit -m "feat: add policy service for registry settings application"
```

---

### Task 7: Chrome Service

**Files:**
- Create: `src/internal/service/chrome.go`
- Create: `src/internal/service/chrome_test.go`

**Interfaces:**
- Consumes: `config.ChromeConfig`, `registry.Registry` (interface)
- Produces: `ChromeService` struct, `NewChromeService()`, `ChromeService.Install()`

- [ ] **Step 1: Write the failing test**

```go
package service

import (
	"context"
	"testing"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

func TestChromeServiceInstall(t *testing.T) {
	mock := &mockRegistry{}
	cfg := config.ChromeConfig{
		ExtensionID:  "abcdefghijklmnop",
		UpdateURL:    "https://clients2.google.com/service/update2/crx",
		ForceInstall: true,
	}

	svc := NewChromeService(cfg, mock)
	err := svc.Install(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mock.stringCalls) != 1 {
		t.Fatalf("expected 1 string call, got %d", len(mock.stringCalls))
	}
	expectedKey := `SOFTWARE\Policies\Google\Chrome\ExtensionInstallForcelist`
	if mock.stringCalls[0].key != expectedKey {
		t.Fatalf("expected key %s, got %s", expectedKey, mock.stringCalls[0].key)
	}
	if mock.stringCalls[0].name != "1" {
		t.Fatalf("expected name '1', got %s", mock.stringCalls[0].name)
	}
	expectedValue := "abcdefghijklmnop;https://clients2.google.com/service/update2/crx"
	if mock.stringCalls[0].value != expectedValue {
		t.Fatalf("expected value %s, got %s", expectedValue, mock.stringCalls[0].value)
	}
}

func TestChromeServiceAlreadyInstalled(t *testing.T) {
	mock := &mockRegistry{}
	mock.valueExists = true
	cfg := config.ChromeConfig{
		ExtensionID:  "abcdefghijklmnop",
		UpdateURL:    "https://clients2.google.com/service/update2/crx",
		ForceInstall: true,
	}

	svc := NewChromeService(cfg, mock)
	err := svc.Install(context.Background())

	if err != nil {
		t.Fatalf("expected no error for already installed, got %v", err)
	}
	if len(mock.stringCalls) != 0 {
		t.Fatalf("expected 0 string calls when already installed, got %d", len(mock.stringCalls))
	}
}
```

But first, the mock registry needs `valueExists` support. Add to `mockRegistry` in `policy_test.go`:

```go
// Add these fields to the mockRegistry struct in policy_test.go:
valueExists bool
existsErr   error

// Update the ValueExists method:
func (m *mockRegistry) ValueExists(key, name string) (bool, error) {
	return m.valueExists, m.existsErr
}
```

- [ ] **Step 2: Update mockRegistry and run test to verify failure**

After adding `valueExists` to mock, run: `cd src && go test ./internal/service/ -v -run TestChrome`
Expected: compilation error — `ChromeService` not defined

- [ ] **Step 3: Write chrome.go**

```go
package service

import (
	"context"
	"fmt"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

const chromeForceInstallKey = `SOFTWARE\Policies\Google\Chrome\ExtensionInstallForcelist`

type ChromeService struct {
	config config.ChromeConfig
	reg    registry.Registry
}

func NewChromeService(cfg config.ChromeConfig, reg registry.Registry) *ChromeService {
	return &ChromeService{config: cfg, reg: reg}
}

func (s *ChromeService) Install(ctx context.Context) error {
	exists, err := s.reg.ValueExists(chromeForceInstallKey, "1")
	if err != nil {
		return fmt.Errorf("chrome: check existing install: %w", err)
	}
	if exists {
		return nil
	}

	value := fmt.Sprintf("%s;%s", s.config.ExtensionID, s.config.UpdateURL)
	if err := s.reg.SetString(chromeForceInstallKey, "1", value); err != nil {
		return fmt.Errorf("chrome: force install extension %s: %w", s.config.ExtensionID, err)
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/service/ -v -run TestChrome`
Expected: both tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/service/chrome.go src/internal/service/chrome_test.go src/internal/service/policy_test.go
git commit -m "feat: add chrome service for extension force-install via registry"
```

---

### Task 8: Desktop Service

**Files:**
- Create: `src/internal/service/desktop.go`
- Create: `src/internal/service/desktop_test.go`

**Interfaces:**
- Consumes: `config.DesktopConfig`, `download.Downloader`, `syscmd.CommandRunner`
- Produces: `DesktopService` struct, `NewDesktopService()`, `DesktopService.Run()`

- [ ] **Step 1: Write the failing test**

```go
package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"bots-util/internal/config"
	"bots-util/internal/pkg/syscmd"
)

type mockDownloader struct {
	downloadFunc func(ctx context.Context, url, dest string, cb func(int64, int64)) error
}

func (m *mockDownloader) Download(ctx context.Context, url, dest string, cb func(int64, int64)) error {
	return m.downloadFunc(ctx, url, dest, cb)
}

type mockCommandRunner struct {
	runFunc func(ctx context.Context, cmd syscmd.Command, cb func(string)) error
}

func (m *mockCommandRunner) Run(ctx context.Context, cmd syscmd.Command, cb func(string)) error {
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/service/ -v -run TestDesktop`
Expected: compilation error — `DesktopService` not defined

- [ ] **Step 3: Write desktop.go**

```go
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
			pct := 0.05 + 0.65*(float64(bytes)/float64(total))
			progress(fmt.Sprintf("downloading... %d/%d bytes", bytes, total), pct)
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/service/ -v -run TestDesktop`
Expected: both tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/service/desktop.go src/internal/service/desktop_test.go
git commit -m "feat: add desktop service for folder cleanup and zip restore"
```

---

### Task 9: Repair Service

**Files:**
- Create: `src/internal/service/repair.go`
- Create: `src/internal/service/repair_test.go`

**Interfaces:**
- Consumes: `config.RepairConfig`, `syscmd.CommandRunner`
- Produces: `RepairService` struct, `NewRepairService()`, `RepairService.RunAll()`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/service/ -v -run TestRepair`
Expected: compilation error — `RepairService` not defined

- [ ] **Step 3: Write repair.go**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/service/ -v -run TestRepair`
Expected: all 3 tests pass

- [ ] **Step 5: Commit**

```bash
git add src/internal/service/repair.go src/internal/service/repair_test.go
git commit -m "feat: add repair service for DISM/SFC orchestration"
```

---

### Task 10: Handler + App Wiring

**Files:**
- Create: `src/internal/handler/handler.go`
- Create: `src/internal/handler/handler_test.go`
- Modify: `src/app.go` (create)
- Modify: `src/main.go` (replace stub)

**Interfaces:**
- Consumes: all four services, Wails v2 runtime
- Produces: `Handler` struct, `App` struct, `main()` entry point

- [ ] **Step 1: Write handler_test.go**

```go
package handler

import (
	"context"
	"errors"
	"testing"
)

type mockDesktopService struct {
	runFunc func(ctx context.Context, progress func(string, float64)) error
}

func (m *mockDesktopService) Run(ctx context.Context, progress func(string, float64)) error {
	if m.runFunc != nil {
		return m.runFunc(ctx, progress)
	}
	return nil
}

type mockChromeService struct {
	installFunc func(ctx context.Context) error
}

func (m *mockChromeService) Install(ctx context.Context) error {
	if m.installFunc != nil {
		return m.installFunc(ctx)
	}
	return nil
}

type mockRepairService struct {
	runAllFunc func(ctx context.Context, progress func(string, float64)) error
}

func (m *mockRepairService) RunAll(ctx context.Context, progress func(string, float64)) error {
	if m.runAllFunc != nil {
		return m.runAllFunc(ctx, progress)
	}
	return nil
}

type mockPolicyService struct {
	applyFunc func(ctx context.Context, progress func(string, int)) error
}

func (m *mockPolicyService) Apply(ctx context.Context, progress func(string, int)) error {
	if m.applyFunc != nil {
		return m.applyFunc(ctx, progress)
	}
	return nil
}

func TestHandlerRunDesktopCleanup(t *testing.T) {
	called := false
	desktop := &mockDesktopService{
		runFunc: func(ctx context.Context, progress func(string, float64)) error {
			called = true
			progress("cleaning", 0.5)
			return nil
		},
	}

	h := NewHandler(desktop, nil, nil, nil)
	err := h.RunDesktopCleanup()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected desktop service to be called")
	}
}

func TestHandlerRunDesktopCleanupError(t *testing.T) {
	desktop := &mockDesktopService{
		runFunc: func(ctx context.Context, progress func(string, float64)) error {
			return errors.New("cleanup failed")
		},
	}

	h := NewHandler(desktop, nil, nil, nil)
	err := h.RunDesktopCleanup()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandlerRunChromeInstall(t *testing.T) {
	called := false
	chrome := &mockChromeService{
		installFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	h := NewHandler(nil, chrome, nil, nil)
	err := h.RunChromeInstall()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected chrome service to be called")
	}
}

func TestHandlerRunSystemRepair(t *testing.T) {
	called := false
	repair := &mockRepairService{
		runAllFunc: func(ctx context.Context, progress func(string, float64)) error {
			called = true
			progress("repairing", 0.5)
			return nil
		},
	}

	h := NewHandler(nil, nil, repair, nil)
	err := h.RunSystemRepair()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected repair service to be called")
	}
}

func TestHandlerApplyPolicySettings(t *testing.T) {
	called := false
	policy := &mockPolicyService{
		applyFunc: func(ctx context.Context, progress func(string, int)) error {
			called = true
			progress("setting policy", 0)
			return nil
		},
	}

	h := NewHandler(nil, nil, nil, policy)
	err := h.ApplyPolicySettings()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected policy service to be called")
	}
}

func TestHandlerAllServices(t *testing.T) {
	h := NewHandler(
		&mockDesktopService{},
		&mockChromeService{},
		&mockRepairService{},
		&mockPolicyService{},
	)

	if err := h.RunDesktopCleanup(); err != nil {
		t.Fatalf("desktop: %v", err)
	}
	if err := h.RunChromeInstall(); err != nil {
		t.Fatalf("chrome: %v", err)
	}
	if err := h.RunSystemRepair(); err != nil {
		t.Fatalf("repair: %v", err)
	}
	if err := h.ApplyPolicySettings(); err != nil {
		t.Fatalf("policy: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd src && go test ./internal/handler/ -v`
Expected: compilation error — `Handler` not defined

- [ ] **Step 3: Write handler.go**

```go
package handler

import (
	"context"
)

type DesktopService interface {
	Run(ctx context.Context, progress func(string, float64)) error
}

type ChromeService interface {
	Install(ctx context.Context) error
}

type RepairService interface {
	RunAll(ctx context.Context, progress func(string, float64)) error
}

type PolicyService interface {
	Apply(ctx context.Context, progress func(string, int)) error
}

type Handler struct {
	desktop DesktopService
	chrome  ChromeService
	repair  RepairService
	policy  PolicyService
}

func NewHandler(d DesktopService, c ChromeService, r RepairService, p PolicyService) *Handler {
	return &Handler{
		desktop: d,
		chrome:  c,
		repair:  r,
		policy:  p,
	}
}

func (h *Handler) RunDesktopCleanup() error {
	ctx := context.Background()
	return h.desktop.Run(ctx, func(msg string, pct float64) {
		// TODO: emit via Wails runtime.EventsEmit in app.go wiring
		_ = msg
		_ = pct
	})
}

func (h *Handler) RunChromeInstall() error {
	ctx := context.Background()
	return h.chrome.Install(ctx)
}

func (h *Handler) RunSystemRepair() error {
	ctx := context.Background()
	return h.repair.RunAll(ctx, func(msg string, pct float64) {
		_ = msg
		_ = pct
	})
}

func (h *Handler) ApplyPolicySettings() error {
	ctx := context.Background()
	return h.policy.Apply(ctx, func(msg string, idx int) {
		_ = msg
		_ = idx
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src && go test ./internal/handler/ -v`
Expected: all 6 tests pass

- [ ] **Step 5: Write app.go**

```go
package main

import (
	"context"

	"bots-util/internal/config"
	"bots-util/internal/handler"
	"bots-util/internal/pkg/download"
	"bots-util/internal/pkg/registry"
	"bots-util/internal/pkg/syscmd"
	"bots-util/internal/service"
)

type App struct {
	ctx     context.Context
	handler *handler.Handler
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	reg := registry.New()
	runner := syscmd.NewRunner()
	downloader := download.NewClient()

	desktopSvc := service.NewDesktopService(
		config.DefaultDesktopConfig(),
		downloader,
		runner,
	)
	chromeSvc := service.NewChromeService(
		config.DefaultChromeConfig(),
		reg,
	)
	repairSvc := service.NewRepairService(
		config.DefaultRepairConfig(),
		runner,
	)
	policySvc := service.NewPolicyService(
		config.DefaultPolicyConfig(),
		reg,
	)

	a.handler = handler.NewHandler(desktopSvc, chromeSvc, repairSvc, policySvc)
}

func (a *App) Shutdown(ctx context.Context) {
}

func (a *App) RunDesktopCleanup() error {
	return a.handler.RunDesktopCleanup()
}

func (a *App) RunChromeInstall() error {
	return a.handler.RunChromeInstall()
}

func (a *App) RunSystemRepair() error {
	return a.handler.RunSystemRepair()
}

func (a *App) ApplyPolicySettings() error {
	return a.handler.ApplyPolicySettings()
}
```

- [ ] **Step 6: Write main.go**

```go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "bots-util",
		Width:  900,
		Height: 650,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 7: Verify the project compiles**

Run: `cd src && go build ./...`
Expected: all packages compile

- [ ] **Step 8: Commit**

```bash
git add src/internal/handler/ src/app.go src/main.go
git commit -m "feat: add handler, app wiring, and Wails entry point"
```

---

### Task 11: Frontend UI

**Files:**
- Create: `src/frontend/src/App.svelte`
- Create: `src/frontend/src/components/Dashboard.svelte`
- Create: `src/frontend/src/components/ActionCard.svelte`
- Create: `src/frontend/src/components/ProgressLog.svelte`

**Interfaces:**
- Consumes: Wails-generated Go bindings (`window.go.main.App.*`)
- Produces: functional dashboard with four independent action cards

- [ ] **Step 1: Write ProgressLog.svelte**

```svelte
<script>
  export let lines = []

  let container

  $: if (container) {
    container.scrollTop = container.scrollHeight
  }
</script>

<div class="progress-log" bind:this={container}>
  {#each lines as line}
    <div class="log-line">{line}</div>
  {/each}
</div>

<style>
  .progress-log {
    background: #0d0d1a;
    border-radius: 6px;
    padding: 10px 14px;
    max-height: 200px;
    overflow-y: auto;
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.6;
  }
  .log-line {
    color: #8be9fd;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
```

- [ ] **Step 2: Write ActionCard.svelte**

```svelte
<script>
  export let title = ''
  export let description = ''
  export let icon = '⚙'
  export let dangerous = false
  export let action = null

  let state = 'idle' // idle | running | success | error | cancelled
  let logLines = []
  let percent = 0
  let errorMessage = ''

  function addLog(msg) {
    logLines = [...logLines, msg]
  }

  async function handleRun() {
    state = 'running'
    logLines = []
    percent = 0
    errorMessage = ''

    window.go.main.App.Log = (msg) => addLog(msg)

    try {
      await action()
      state = 'success'
      percent = 100
    } catch (e) {
      state = 'error'
      errorMessage = e.message || String(e)
      addLog('ERROR: ' + errorMessage)
    }
  }

  function handleCancel() {
    state = 'cancelled'
    addLog('Cancelled by user')
  }

  $: statusClass = state === 'running' ? 'running' :
                   state === 'success' ? 'success' :
                   state === 'error' ? 'error' :
                   state === 'cancelled' ? 'cancelled' : ''
</script>

<div class="card {statusClass}">
  <div class="card-header">
    <span class="icon">{icon}</span>
    <div class="header-text">
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  </div>

  {#if state === 'running'}
    <div class="progress-bar">
      <div class="progress-fill" style="width: {percent}%"></div>
    </div>
  {/if}

  {#if logLines.length > 0}
    <ProgressLog lines={logLines} />
  {/if}

  {#if state === 'error'}
    <div class="error-banner">{errorMessage}</div>
  {/if}

  <div class="card-actions">
    {#if state === 'idle' || state === 'success' || state === 'error' || state === 'cancelled'}
      <button
        class="btn {dangerous ? 'btn-danger' : 'btn-primary'}"
        on:click={handleRun}
      >
        {state === 'idle' ? 'Run' : 'Run Again'}
      </button>
    {/if}
    {#if state === 'running'}
      <button class="btn btn-cancel" on:click={handleCancel}>
        Cancel
      </button>
    {/if}
  </div>
</div>

<style>
  .card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 10px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    transition: border-color 0.2s;
  }
  .card.running { border-color: #e6a817; }
  .card.success { border-color: #2ecc71; }
  .card.error { border-color: #e74c3c; }
  .card.cancelled { border-color: #7f8c8d; }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }
  .icon { font-size: 24px; }
  .header-text h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }
  .header-text p {
    margin: 4px 0 0;
    font-size: 13px;
    color: #8892b0;
  }

  .progress-bar {
    background: #0d0d1a;
    border-radius: 4px;
    height: 6px;
    overflow: hidden;
  }
  .progress-fill {
    background: #e6a817;
    height: 100%;
    transition: width 0.3s;
  }

  .card-actions {
    display: flex;
    gap: 8px;
  }
  .btn {
    padding: 8px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .btn-primary { background: #0f3460; color: #e0e0e0; }
  .btn-primary:hover { background: #1a4a7a; }
  .btn-danger { background: #c0392b; color: #e0e0e0; }
  .btn-danger:hover { background: #e74c3c; }
  .btn-cancel { background: #7f8c8d; color: #e0e0e0; }
  .btn-cancel:hover { background: #95a5a6; }

  .error-banner {
    background: rgba(231, 76, 60, 0.15);
    color: #e74c3c;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
  }
</style>
```

- [ ] **Step 3: Write Dashboard.svelte**

```svelte
<script>
  import ActionCard from './ActionCard.svelte'

  let actions = [
    {
      title: 'Desktop Cleanup',
      description: 'Clear and restore the desktop folder from a zip archive',
      icon: '🗂',
      dangerous: false,
      action: () => window.go.main.App.RunDesktopCleanup(),
    },
    {
      title: 'Chrome Extension',
      description: 'Force-install the Chrome extension via registry policy',
      icon: '🧩',
      dangerous: false,
      action: () => window.go.main.App.RunChromeInstall(),
    },
    {
      title: 'System Repair',
      description: 'Run DISM, SFC, and other system repair utilities',
      icon: '🔧',
      dangerous: true,
      action: () => window.go.main.App.RunSystemRepair(),
    },
    {
      title: 'Policy Settings',
      description: 'Apply group policy and registry settings',
      icon: '📋',
      dangerous: true,
      action: () => window.go.main.App.ApplyPolicySettings(),
    },
  ]
</script>

<div class="dashboard">
  <header class="header">
    <h1>bots-util</h1>
    <span class="subtitle">Desktop Utility</span>
  </header>
  <div class="grid">
    {#each actions as action}
      <ActionCard
        title={action.title}
        description={action.description}
        icon={action.icon}
        dangerous={action.dangerous}
        action={action.action}
      />
    {/each}
  </div>
</div>

<style>
  .dashboard {
    max-width: 900px;
    margin: 0 auto;
  }
  .header {
    text-align: center;
    margin-bottom: 32px;
  }
  .header h1 {
    font-size: 28px;
    font-weight: 700;
    color: #e0e0e0;
  }
  .subtitle {
    font-size: 14px;
    color: #8892b0;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
    gap: 20px;
  }
</style>
```

- [ ] **Step 4: Write App.svelte**

```svelte
<script>
  import Dashboard from './components/Dashboard.svelte'
</script>

<Dashboard />
```

- [ ] **Step 5: Install frontend dependencies**

Run: `cd src/frontend && npm install`

- [ ] **Step 6: Verify frontend builds**

Run: `cd src/frontend && npm run build`
Expected: builds without errors, produces `frontend/dist/`

- [ ] **Step 7: Commit**

```bash
git add src/frontend/src/ src/frontend/package-lock.json
git commit -m "feat: add Svelte dashboard with four action cards and progress logs"
```

---

### Task 12: Integration Verification

**Files:**
- None (verification only)

**Interfaces:**
- Consumes: entire project
- Produces: confirmed working build

- [ ] **Step 1: Verify full Go project compiles**

Run: `cd src && go build ./...`
Expected: all packages compile without errors

- [ ] **Step 2: Run all Go unit tests**

Run: `cd src && go test ./...`
Expected: all tests pass (registry integration tests may be skipped on macOS)

- [ ] **Step 3: Verify frontend builds**

Run: `cd src/frontend && npm run build`
Expected: builds without errors

- [ ] **Step 4: Build Wails application for Windows** (on Windows machine)

Run: `cd src && wails build`
Expected: produces `build/bin/bots-util.exe`

- [ ] **Step 5: Manual smoke test on Windows VM/device**

Verify:
- App launches and shows 4 action cards
- Each card shows its Run button
- "System Repair" and "Policy Settings" marked as dangerous (red styling)
- Run buttons trigger actions (verify via logs/registry changes)
- Cancel buttons work during long-running operations
- Error states display correctly

- [ ] **Step 6: Commit any final adjustments**

```bash
git add -A
git commit -m "chore: final integration verification and cleanup"
```