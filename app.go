package main

import (
	"embed"
	"log"

	"quick-translate/internal/transport"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

// Runs the application with the adapter registered as the only bridge between the frontend and the application. Blocks until the application is stopped and ends the process if it can not be started.
func runApp(adapter *transport.Adapter) {
	err := wails.Run(&options.App{
		Title:             "Quick Translate",
		Width:             400,
		Height:            250,
		DisableResize:     true,
		Frameless:         true,
		StartHidden:       true,
		AlwaysOnTop:       true,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        adapter.StartUp,
		OnShutdown:       adapter.Shutdown,
		Bind: []interface{}{
			adapter,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: true,
		},
	})

	if err != nil {
		log.Fatalf("Could not start the application: %v", err)
	}
}
