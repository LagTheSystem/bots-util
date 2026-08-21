package service

import (
	"context"
	"fmt"

	"bots-util/internal/config"
	"bots-util/internal/pkg/registry"
)

const chromeForceInstallKey = `SOFTWARE\Policies\Google\Chrome\ExtensionInstallForcelist`

type ChromeService struct {
	config config.ChromeConfig
	reg    registry.Registry
}

func NewChromeService(cfg config.ChromeConfig, reg registry.Registry) *ChromeService {
	return &ChromeService{config: cfg, reg: reg}
}

func (s *ChromeService) Install(ctx context.Context) error {
	exists, err := s.reg.ValueExists(chromeForceInstallKey, "1")
	if err != nil {
		return fmt.Errorf("chrome: check existing install: %w", err)
	}
	if exists {
		return nil
	}

	value := fmt.Sprintf("%s;%s", s.config.ExtensionID, s.config.UpdateURL)
	if err := s.reg.SetString(chromeForceInstallKey, "1", value); err != nil {
		return fmt.Errorf("chrome: force install extension %s: %w", s.config.ExtensionID, err)
	}

	return nil
}
