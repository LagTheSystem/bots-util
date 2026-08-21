package service

import (
	"context"
	"testing"

	"bots-util/internal/config"
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
