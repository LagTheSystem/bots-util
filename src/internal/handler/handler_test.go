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
	applyFunc func(ctx context.Context, progress func(string, float64)) error
}

func (m *mockPolicyService) Apply(ctx context.Context, progress func(string, float64)) error {
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
		applyFunc: func(ctx context.Context, progress func(string, float64)) error {
			called = true
			progress("setting policy", 0.0)
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
