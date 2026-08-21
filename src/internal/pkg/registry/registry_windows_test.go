//go:build windows

package registry

import (
	"testing"
)

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