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
