//go:build windows

package registry

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

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
