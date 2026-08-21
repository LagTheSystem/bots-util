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
	valueExists  bool
	existsErr    error
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
	return m.valueExists, m.existsErr
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
