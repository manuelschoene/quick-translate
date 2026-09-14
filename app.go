package main

import (
	"embed"

	"quick-translate/internal/transport"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed art/quick-translate_48x48.png
var icon []byte

// Runs the application with the adapter registered as the only bridge between the frontend and the application. Blocks until the application is stopped and ends the process if it can not be started.
func runApp() {
	app := application.New(application.Options{
		Name:        "Quick Translate",
		Description: "Translate your selected text anywhere",
		Icon:        icon,
		Linux: application.LinuxOptions{
			ProgramName: "quick-translate",
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.RegisterService(application.NewService(transport.NewAdapter(app)))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Quick Translate",
		Width:            480,
		Height:           300,
		AlwaysOnTop:      true,
		DisableResize:    true,
		Frameless:        true,
		BackgroundColour: application.RGBA{Red: 1, Green: 3, Blue: 3, Alpha: 255},
		Hidden:           !isDevBuild(),
		Linux: application.LinuxWindow{
			Icon:             icon,
			WebviewGpuPolicy: application.WebviewGpuPolicyAlways,
		},
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
