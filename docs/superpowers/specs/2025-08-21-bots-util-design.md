# bots-util Design Spec

## Overview

**bots-util** — a Windows desktop utility for technician and end-user machine configuration. Built with Go (backend) + Wails (desktop shell) + Svelte (UI). Provides four independent, individually-triggerable actions from a dashboard.

## Goals

- **Clarity for both audiences:** IT staff see detailed logs; end users see readable status and progress
- **Independent actions:** Any action can run alone or alongside others; each manages its own state
- **Hardcoded configuration:** No external config files — all paths, URLs, extension IDs, and registry values live in `internal/config/config.go`
- **Windows-only:** All system dependencies (registry, DISM, SFC, GPO) are Windows-native
- **Testable business logic:** Services behind interfaces, shared packages injectable as mocks

## Package Tree

```
src/
├── main.go                      # Wails entry point: creates App, starts Wails
├── app.go                       # App struct (Startup/Shutdown hooks; initializes services)
├── wails.json                   # Wails build config (name, assetdir, etc.)
├── go.mod / go.sum
├── internal/
│   ├── config/
│   │   └── config.go            # All hardcoded constants in typed structs
│   ├── handler/
│   │   └── handler.go           # 4 Wails-bound methods, delegates to services
│   ├── service/
│   │   ├── desktop.go           # Clear folder + download & extract zip
│   │   ├── chrome.go            # Registry force-install of Chrome extension
│   │   ├── repair.go            # DISM, SFC, chkdsk — run in sequence
│   │   └── policy.go            # Group policy / registry settings writes
│   └── pkg/
│       ├── registry/
│       │   └── registry.go      # Windows registry read/write/delete
│       ├── syscmd/
│       │   └── syscmd.go        # Run commands with timeout, cancellation, streaming
│       └── download/
│           └── download.go      # HTTP download with progress callback
├── frontend/
│   ├── index.html
│   ├── package.json
│   └── src/
│       ├── main.js
│       ├── App.svelte
│       └── components/
│           ├── Dashboard.svelte
│           ├── ActionCard.svelte
│           └── ProgressLog.svelte
```

## Component Specifications

### `internal/config/config.go`

One package-level variable or function exposing typed config. Each feature gets its own struct:

```go
type DesktopConfig struct {
    TargetFolder string // e.g. C:\Users\Public\Desktop\CampShortcuts
    ZipURL       string
    TempDir      string // where to download zip before extraction
}

type ChromeConfig struct {
    ExtensionID   string // e.g. abcdefghijklmnop
    UpdateURL     string
    ForceInstall  bool
}

type RepairConfig struct {
    Commands []RepairCommand // DISM, SFC, etc. in order
}

type RepairCommand struct {
    Name string
    Exe  string
    Args []string
}

type PolicyConfig struct {
    RegistryEntries []RegistryEntry
}

type RegistryEntry struct {
    KeyPath string // HKLM\Software\Policies\...
    Name    string
    Value   interface{}
    Type    uint32 // REG_DWORD, REG_SZ, etc.
}
```

Config is initialized once at startup, passed to service constructors. No hot-reload — this is compiled in.

### `internal/pkg/registry/registry.go`

Wraps `golang.org/x/sys/windows/registry`. Exposes:

```go
type Registry struct{}

func New() *Registry
func (r *Registry) SetDWORD(key, name string, value uint32) error
func (r *Registry) SetString(key, name, value string) error
func (r *Registry) Delete(key, name string) error
func (r *Registry) KeyExists(key string) (bool, error)
func (r *Registry) ValueExists(key, name string) (bool, error)
```

All methods take full registry paths (e.g., `SOFTWARE\Policies\Google\Chrome`). The package translates to the `registry.Key` value internally.

**Testing:** Integration tests on real Windows (can't mock win32 registry API meaningfully).

### `internal/pkg/syscmd/syscmd.go`

```go
type LineCallback func(line string)
type Command struct { Name, Exe string; Args []string; Timeout time.Duration }

type Runner struct{}

func NewRunner() *Runner
func (r *Runner) Run(ctx context.Context, cmd Command, cb LineCallback) error
```

- `Run` creates an `exec.Command`, pipes stdout/stderr, reads line by line, calls `cb` for each line
- Respects `ctx.Done()` — sends `os.Kill` on cancellation
- Respects `cmd.Timeout` — auto-cancels after timeout
- Returns combined error if command exits non-zero

**Testing:** Unit tests run real commands (e.g., `cmd /c echo hello`, `cmd /c exit 1`).

### `internal/pkg/download/download.go`

```go
type ProgressCallback func(bytesDownloaded, totalBytes int64)
type Client struct { http.Client }

func NewClient() *Client
func (c *Client) Download(ctx context.Context, url, dest string, cb ProgressCallback) error
```

- Uses `net/http` with `context.Context` for cancellation
- If `totalBytes` is unknown (no Content-Length), passes -1
- Writes to file on disk, not memory
- Respects `ctx.Done()` — tears down connection

**Testing:** Unit tests with `httptest.Server`.

### `internal/service/desktop.go`

```go
type DesktopService struct { config config.DesktopConfig; dl *download.Client; sc *syscmd.Runner }
func NewDesktopService(c config.DesktopConfig, dl *download.Client, sc *syscmd.Runner) *DesktopService
func (s *DesktopService) Run(ctx context.Context, progress func(string, float64)) error
```

**Algorithm:**
1. Delete all contents of `TargetFolder` (not the folder itself)
2. Download zip from `ZipURL` to temp location (report bytes as progress 0.0–0.7)
3. Extract zip into `TargetFolder` using system `tar` or PowerShell `Expand-Archive` (report 0.7–1.0)
4. Clean up temp zip

**Error handling:** If folder doesn't exist, create it. If zip download fails, leave folder empty and report error. If extraction fails mid-way, report error and clean up.

**Testing:** Inject mock `download.Client` and `syscmd.Runner`.

### `internal/service/chrome.go`

```go
type ChromeService struct { config config.ChromeConfig; reg *registry.Registry }
func NewChromeService(c config.ChromeConfig, reg *registry.Registry) *ChromeService
func (s *ChromeService) Install(ctx context.Context) error
```

**Algorithm:**
1. Write the `ExtensionInstallForcelist` registry key:
   - Key: `HKLM\SOFTWARE\Policies\Google\Chrome\ExtensionInstallForcelist`
   - Value name: numeric index (e.g., `1`)
   - Value: `extensionID;updateURL`
2. Log what was written
3. If already present, log "already installed" and return nil (not an error)

**Testing:** Inject mock `Registry`.

### `internal/service/repair.go`

```go
type RepairService struct { config config.RepairConfig; sc *syscmd.Runner }
func NewRepairService(c config.RepairConfig, sc *syscmd.Runner) *RepairService
func (s *RepairService) RunAll(ctx context.Context, progress func(string, float64)) error
```

**Algorithm:**
1. Run each `RepairCommand` in order
2. Report progress as `currentIndex / total`, with each command's output lines streamed via the `progress` callback
3. Default command set: `DISM /Online /Cleanup-Image /RestoreHealth` → `SFC /ScanNow` → optional `chkdsk /f`
4. If a command fails, log the error and continue to the next one (fail-soft)
5. Return aggregated success/failure

**Testing:** Inject mock `syscmd.Runner`.

### `internal/service/policy.go`

```go
type PolicyService struct { config config.PolicyConfig; reg *registry.Registry }
func NewPolicyService(c config.PolicyConfig, reg *registry.Registry) *PolicyService
func (s *PolicyService) Apply(ctx context.Context, progress func(string, int)) error
```

**Algorithm:**
1. Iterate over `RegistryEntries`
2. Write each using `registry.SetDWORD` or `registry.SetString`
3. Log each write
4. Report progress as `i / len(entries)`
5. Return nil if all succeed; if any fail, return aggregated error

**Testing:** Inject mock `Registry`.

### `internal/handler/handler.go`

```go
type Handler struct {
    desktop *service.DesktopService
    chrome  *service.ChromeService
    repair  *service.RepairService
    policy  *service.PolicyService
}

func NewHandler(d, c, r, p interface{ /* service interfaces */ }) *Handler

// Each method is a Wails runtime-bound function:
func (h *Handler) RunDesktopCleanup() error
func (h *Handler) RunChromeInstall() error
func (h *Handler) RunSystemRepair() error
func (h *Handler) ApplyPolicySettings() error
```

Each method:
1. Creates a `context.WithCancel` and stores cancel for potential stop
2. Wraps the service's progress callback into a Wails `runtime.EventsEmit` call to stream to frontend
3. Returns error or nil
4. Handles cancellation via a separate `Cancel*` method

The handler is intentionally thin — 4 methods, ~10 lines each.

### `app.go`

```go
type App struct {
    ctx     context.Context
    handler *handler.Handler
}

func (a *App) Startup(ctx context.Context) { /* initialize services, pass to handler */ }
func (a *App) Shutdown(ctx context.Context) { /* cancel any running operations */ }
```

## Frontend (Svelte)

**Dashboard.svelte:** Grid of 4 ActionCards. Header with app title. Global error banner at top.

**ActionCard.svelte:**
- Props: `title`, `description`, `icon`, `dangerous` (boolean for styling)
- States: `idle | running | success | error | cancelled`
- Shows run button in `idle`, cancel button in `running`, status icon in `success`/`error`/`cancelled`
- Progress bar + inline log when `running` or in terminal state

**ProgressLog.svelte:**
- Scrollable terminal-style log window
- Each line timestamped or color-coded (stdout vs stderr vs status message)
- Auto-scrolls to bottom on new lines

All Wails bindings called via generated JS API (Wails v3 `Call` pattern or `@wails/runtime` equivalent).

## Data Flow

```
User clicks "Run Repair"
  │
  ▼
Svelte ActionCard ──call──▶ Wails Runtime ──invoke──▶ handler.RunSystemRepair()
                                                            │
                                                            ▼
                                                    service.RepairService
                                                            │
                                                     creates ctx, cancel
                                                            │
                                              ┌─────────────┼─────────────┐
                                              ▼             ▼             ▼
                                          DISM cmd      SFC cmd      chkdsk cmd
                                              │             │             │
                                              └──────┬──────┴──────┬──────┘
                                                     │             │
                                            syscmd.Runner.Run()
                                                     │
                                               streams lines
                                                     │
                                                     ▼
                                            progress callback
                                                     │
                                                     ▼
                                        Wails EventsEmit("progress", ...)
                                                     │
                                                     ▼
                                            Svelte ProgressLog updates
```

## Error Handling

| Layer | Strategy |
|---|---|
| `pkg/syscmd` | Wraps `exec.ExitError` with command name, exit code, stderr. Surface in structured error. |
| `pkg/registry` | Wraps Windows `syscall.Errno` with key path and operation attempted. |
| `pkg/download` | Wraps HTTP errors with URL and status code. File I/O errors with path. |
| `service/*` | Each service catches pkg errors, wraps with context ("failed to download zip", "failed at SFC step"), returns unified error. |
| `handler` | Passes service errors directly to frontend — no additional wrapping. |
| `frontend` | Displays `error.message` in red. Log lines remain visible for debugging. |

## Testing Summary

| Target | Type | Scope |
|---|---|---|
| `pkg/syscmd` | Unit | Run `cmd /c echo`, `cmd /c exit 1`, cancellation via context |
| `pkg/registry` | Integration | Real Windows registry writes in temp key paths |
| `pkg/download` | Unit | `httptest.Server` with known files, errors, slow responses |
| `service/desktop` | Unit | Mock download + mock syscmd |
| `service/chrome` | Unit | Mock registry |
| `service/repair` | Unit | Mock syscmd |
| `service/policy` | Unit | Mock registry |
| `handler` | Unit | Mock all services, verify delegation |
| `frontend` | Manual | MVP scope — no JS test framework |

## Assumptions & Edge Cases

| Scenario | Behavior |
|---|---|
| Folder to clean doesn't exist | Create it, continue |
| Zip download times out | Report error, folder remains empty |
| Chrome extension already force-installed | Log note, return success (idempotent) |
| User cancels mid-repair | Kill running command; subsequent commands DO NOT run |
| Two actions started simultaneously | No locking — services are stateless per-run; concurrent execution is allowed |
| Running DISM requires admin | DISM/SFC/chkdsk all require elevation. The app must be run elevated or self-elevate on repair action |
| No internet for zip download | HTTP error surfaced as friendly message |
| Target folder on network drive | Should work; no special handling |
| Chrome not installed | Registry key written anyway — harmless; takes effect when Chrome is installed |

## Implementation Order

1. Scaffold Wails project (`wails init`) with Svelte template
2. Implement `pkg/syscmd` + tests
3. Implement `pkg/registry` + tests
4. Implement `pkg/download` + tests
5. Implement `service/policy` + tests (simplest service, validates registry pkg)
6. Implement `service/chrome` + tests
7. Implement `service/desktop` + tests
8. Implement `service/repair` + tests
9. Implement `handler.go` + app.go wiring + tests
10. Build Svelte UI components
11. Integrate frontend ↔ backend via Wails bindings
12. Manual integration testing on Windows VM/device
