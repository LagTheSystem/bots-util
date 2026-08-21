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
	emit    func(action, msg string, pct float64)
}

func NewHandler(d DesktopService, c ChromeService, r RepairService, p PolicyService) *Handler {
	return &Handler{
		desktop: d,
		chrome:  c,
		repair:  r,
		policy:  p,
	}
}

func (h *Handler) SetEmitter(emit func(action, msg string, pct float64)) {
	h.emit = emit
}

func (h *Handler) emitProgress(action, msg string, pct float64) {
	if h.emit != nil {
		h.emit(action, msg, pct)
	}
}

func (h *Handler) RunDesktopCleanup() error {
	ctx := context.Background()
	return h.desktop.Run(ctx, func(msg string, pct float64) {
		h.emitProgress("desktop", msg, pct)
	})
}

func (h *Handler) RunChromeInstall() error {
	ctx := context.Background()
	h.emitProgress("chrome", "installing chrome extension", 0)
	if err := h.chrome.Install(ctx); err != nil {
		return err
	}
	h.emitProgress("chrome", "done", 1)
	return nil
}

func (h *Handler) RunSystemRepair() error {
	ctx := context.Background()
	return h.repair.RunAll(ctx, func(msg string, pct float64) {
		h.emitProgress("repair", msg, pct)
	})
}

func (h *Handler) ApplyPolicySettings() error {
	ctx := context.Background()
	return h.policy.Apply(ctx, func(msg string, idx int) {
		h.emitProgress("policy", msg, float64(idx))
	})
}
