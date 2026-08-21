package main

import (
	"context"

	"bots-util/internal/config"
	"bots-util/internal/handler"
	"bots-util/internal/pkg/download"
	"bots-util/internal/pkg/registry"
	"bots-util/internal/pkg/syscmd"
	"bots-util/internal/service"
)

type App struct {
	ctx     context.Context
	handler *handler.Handler
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	reg := registry.New()
	runner := syscmd.NewRunner()
	downloader := download.NewClient()

	desktopSvc := service.NewDesktopService(
		config.DefaultDesktopConfig(),
		downloader,
		runner,
	)
	chromeSvc := service.NewChromeService(
		config.DefaultChromeConfig(),
		reg,
	)
	repairSvc := service.NewRepairService(
		config.DefaultRepairConfig(),
		runner,
	)
	policySvc := service.NewPolicyService(
		config.DefaultPolicyConfig(),
		reg,
	)

	a.handler = handler.NewHandler(desktopSvc, chromeSvc, repairSvc, policySvc)
}

func (a *App) Shutdown(ctx context.Context) {
}

func (a *App) RunDesktopCleanup() error {
	return a.handler.RunDesktopCleanup()
}

func (a *App) RunChromeInstall() error {
	return a.handler.RunChromeInstall()
}

func (a *App) RunSystemRepair() error {
	return a.handler.RunSystemRepair()
}

func (a *App) ApplyPolicySettings() error {
	return a.handler.ApplyPolicySettings()
}
