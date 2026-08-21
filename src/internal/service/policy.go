package service

import (
	"context"
	"fmt"
	"strings"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

type PolicyService struct {
	config config.PolicyConfig
	reg    registry.Registry
}

func NewPolicyService(cfg config.PolicyConfig, reg registry.Registry) *PolicyService {
	return &PolicyService{config: cfg, reg: reg}
}

func (s *PolicyService) Apply(ctx context.Context, progress func(string, int)) error {
	if len(s.config.RegistryEntries) == 0 {
		if progress != nil {
			progress("no registry entries to apply", 0)
		}
		return nil
	}

	var errs []string

	for i, entry := range s.config.RegistryEntries {
		var err error
		switch entry.Type {
		case registry.DWORD:
			val, ok := entry.Value.(uint32)
			if !ok {
				errs = append(errs, fmt.Sprintf("policy: %s: invalid DWORD value type", entry.Name))
				continue
			}
			err = s.reg.SetDWORD(entry.KeyPath, entry.Name, val)
		case registry.SZ:
			val, ok := entry.Value.(string)
			if !ok {
				errs = append(errs, fmt.Sprintf("policy: %s: invalid string value type", entry.Name))
				continue
			}
			err = s.reg.SetString(entry.KeyPath, entry.Name, val)
		default:
			errs = append(errs, fmt.Sprintf("policy: %s: unsupported registry type %d", entry.Name, entry.Type))
			continue
		}

		if err != nil {
			msg := fmt.Sprintf("failed to set %s\\%s: %v", entry.KeyPath, entry.Name, err)
			errs = append(errs, msg)
			if progress != nil {
				progress(msg, i)
			}
		} else {
			msg := fmt.Sprintf("set %s\\%s", entry.KeyPath, entry.Name)
			if progress != nil {
				progress(msg, i)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("policy: %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}
