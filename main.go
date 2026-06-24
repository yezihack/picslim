package main

import (
	"context"
	"embed"
	"log"

	"github.com/yezihack/PicSlim/internal/httpapi"
	"github.com/yezihack/PicSlim/internal/logx"
	"github.com/yezihack/PicSlim/internal/config"

	"github.com/yezihack/PicSlim/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg := config.LoadDefault()
	logger := logx.New(cfg.LogLevel)
	defer func() { _ = logger.Sync() }()

	appInstance := app.New(logger)
	httpServer := httpapi.NewServer(cfg.HTTPPort, logger)

	err := wails.Run(&options.App{
		Title:  "智能图片压缩器",
		Width:  1380,
		Height: 920,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			appInstance.Startup(ctx)
			if cfg.EnableLocalHTTP {
				go httpServer.Start()
			}
		},
		OnShutdown: func(ctx context.Context) {
			_ = ctx
			if cfg.EnableLocalHTTP {
				_ = httpServer.Stop()
			}
		},
		Bind: []any{
			appInstance,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
