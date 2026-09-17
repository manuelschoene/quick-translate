package main

import (
	"embed"

	"quick-translate/internal/desktop"
	"quick-translate/internal/transport"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
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
			// The name everything on the desktop matches on: the desktop entry file, the window's class under
			// Wayland and X11, the autostart entry, the single-instance lock and the KWin rule. ProgramName
			// is left unset on purpose, because Wails then gives it this same id, and only then does a
			// window under Wayland match its desktop entry.
			ApplicationID: desktop.ApplicationID,
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
		Hidden:           !app.Env.Info().Debug,
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
