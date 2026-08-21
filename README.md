# bots-util
[![Build](https://github.com/LagTheSystem/bots-util/actions/workflows/build.yml/badge.svg)](https://github.com/LagTheSystem/bots-util/actions/workflows/build.yml)

Desktop utility for camp setup.
A Windows desktop app built with **Go** (backend) + **Wails v2** (desktop shell) + **Svelte** (UI). It provides four independent actions from a dashboard:

- **Desktop Cleanup** — clear a folder on the Desktop and restore its contents from a downloaded zip
- **Chrome Extension** — force-install a Chrome extension via registry policy
- **System Repair** — run DISM, SFC, and other system repair utilities
- **Policy Settings** — apply group policy and registry settings

## Prerequisites

- **Windows** — this app targets Windows (registry, DISM, SFC, GPO). Build and run it on Windows.
- **Go** — see `src/go.mod` for the required version
- **Wails** CLI — `go install github.com/wailsapp/wails/v2` or `go get` per the Wails docs
- **Node.js / npm** — for building the Svelte frontend

## Build

All commands run from the `src/` directory.

### 1. Install frontend dependencies

```bash
cd src/frontend
npm install
```

### 2. Build the frontend

```bash
cd src/frontend
npm run build
```

This produces `src/frontend/dist/`, which the Go app embeds at compile time.

### 3. Build the Windows executable

```bash
cd src
wails build
```

The executable is written to `src/build/bin/bots-util.exe`.

## Run (development)

```bash
cd src/frontend
npm run dev
```

then launch the app from your IDE or with `wails dev` for hot-reload of the backend and frontend.

## Tests

```bash
cd src
go test ./...
```

The Windows registry integration tests are tagged `//go:build windows` and only run on Windows. On other platforms a stub returns errors so the project still compiles for development.

## Configuration

All configurable values (folder path, zip URL, extension ID, repair commands, registry settings) are hardcoded in `src/internal/config/config.go`. Edit them there and rebuild.
