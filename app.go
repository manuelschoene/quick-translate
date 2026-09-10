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

//go:embed art/quick-translate_48x48.png
var icon []byte

// Runs the application with the adapter registered as the only bridge between the frontend and the application. Blocks until the application is stopped and ends the process if it can not be started.
func runApp(adapter *transport.Adapter) {
	err := wails.Run(&options.App{
		Title:             "Quick Translate",
		Width:             480,
		Height:            300,
		DisableResize:     true,
		Frameless:         true,
		StartHidden:       !isDevBuild(),
		AlwaysOnTop:       true,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 1, G: 3, B: 3, A: 255},
		OnStartup:        adapter.StartUp,
		OnShutdown:       adapter.Shutdown,
		Bind: []interface{}{
			adapter,
		},
		Linux: &linux.Options{
			Icon:        icon,
			ProgramName: "quick-translate",
		},
	})

	if err != nil {
		log.Fatalf("Could not start the application: %v", err)
	}
}
