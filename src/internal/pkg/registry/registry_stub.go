//go:build !windows

package registry

import (
	"fmt"
)

type RegistryImpl struct{}

func New() *RegistryImpl {
	return &RegistryImpl{}
}

func (r *RegistryImpl) SetDWORD(key, name string, value uint32) error {
	return fmt.Errorf("registry: SetDWORD unsupported on this platform")
}

func (r *RegistryImpl) SetString(key, name, value string) error {
	return fmt.Errorf("registry: SetString unsupported on this platform")
}

func (r *RegistryImpl) Delete(key, name string) error {
	return fmt.Errorf("registry: Delete unsupported on this platform")
}

func (r *RegistryImpl) KeyExists(key string) (bool, error) {
	return false, fmt.Errorf("registry: KeyExists unsupported on this platform")
}

func (r *RegistryImpl) ValueExists(key, name string) (bool, error) {
	return false, fmt.Errorf("registry: ValueExists unsupported on this platform")
}