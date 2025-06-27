package service

import (
	"context"
	"speedtest-gd/global"
	"speedtest-gd/runtimes"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	runtime.WindowSetMinSize(ctx, 1100, 600)
}

func (a *App) InitNode() string {
	return initNode()
}

func (a *App) GetInfo() string {
	return getInfo()
}

func (a *App) GetAllNode() string {
	return runtimes.ShowAllNode(global.GlobalApacheAgents)
}

func (a *App) GetBestNode() string {
	return showBestNode()
}

func (a *App) StartTest(id []int, threads int, mode int) {
	startTest(id, threads, mode)
}

func (a *App) PingSelectedNode(id int) string {
	return pingSelectedNode(id)
}

func (a *App) GetSpeed() string {
	return getspeed()
}
