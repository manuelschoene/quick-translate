package main

import (
	"embed"

	"quick-translate/internal/system"
	"quick-translate/internal/transport"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// Runs the Wails application. Registers services and creates the window. No icon is handed over as it is ignored by GTK4 and taken from the desktop entry instead. Blocks until the application is stopped.
func runApp() {
	var adapter *transport.Adapter
	
	app := application.New(application.Options{
		Name:        "Quick Translate",
		Description: "Translate your selected text anywhere",
		Linux: application.LinuxOptions{
			ApplicationID: system.ApplicationID, // Required for application matching with desktop entry, KWin rules, etc.
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: system.ApplicationID, // Owns '<id>.SingleInstance' on the session bus
			OnSecondInstanceLaunch: func(_ application.SecondInstanceData) { adapter.Restore() }, // Same adapter for both instances
		},
	})
	
	files := system.NewFileService()
	adapter = transport.NewAdapter(app, files)
	
	app.RegisterService(application.NewService(system.NewIntegrationService(app, files, adapter.Show)))
	app.RegisterService(application.NewService(adapter))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Quick Translate",
		Width:            480,
		Height:           300,
		AlwaysOnTop:      true,
		DisableResize:    true,
		Frameless:        true,
		BackgroundColour: application.RGBA{Red: 1, Green: 3, Blue: 3, Alpha: 255},
		Hidden:           !app.Env.Info().Debug,
		Linux: application.LinuxWindow{
			WebviewGpuPolicy: application.WebviewGpuPolicyAlways,
		},
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
