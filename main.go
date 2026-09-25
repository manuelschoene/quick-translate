package main

import (
	"embed"
	"log"
	"os"

	"quick-translate/internal/cli"
	"quick-translate/internal/system"
	"quick-translate/internal/transport"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// The version this binary was built as. Stamped in by the build from the version in 'build/config.yml'. Development builds use this default.
var version = "0.0.0-dev"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed art/tray.png
var trayIcon []byte

// Starts the application. A command given on the command line is carried out instead, before anything of the application is set up, so a command is never handed to the instance that is already running.
func main() {
	if handled, code := cli.Run(os.Args[1:], version); handled {
		os.Exit(code)
	}

	var adapter *transport.Adapter

	// Built here rather than inside createApp, so it closes over this variable and sees the adapter wireServices returns.
	app := createApp(func() {
		// Nil only between application.New() taking the lock and the wiring below.
		if adapter != nil {
			adapter.Restore()
		}
	})

	registerWindows(app)

	adapter = wireServices(app)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// Creates the application with its options, but does not yet register any services or windows. The second instance callback is handed in, because the lock it answers is taken in here, before anything it could reach exists.
func createApp(onSecondInstance func()) *application.App {
	return application.New(application.Options{
		Name:        "Quick Translate",
		Description: "Translate your selected text anywhere",
		Linux: application.LinuxOptions{
			ApplicationID: system.ApplicationID, // Required for application matching with desktop entry, KWin rules, etc.
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:               system.ApplicationID, // Owns '<id>.SingleInstance' on the session bus
			OnSecondInstanceLaunch: func(_ application.SecondInstanceData) { onSecondInstance() },
		},
	})
}

// Registers the main window of the application, which is a frameless webview window that is always on top and cannot be resized. The window is hidden in production builds, but shown in debug builds.
func registerWindows(app *application.App) {
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Quick Translate",
		Width:            480,
		Height:           300,
		AlwaysOnTop:      true,
		DisableResize:    true,
		Frameless:        true,
		BackgroundColour: application.NewRGBA(1, 3, 3, 255),
		Hidden:           !app.Env.Info().Debug,
		Linux: application.LinuxWindow{
			WebviewGpuPolicy: application.WebviewGpuPolicyAlways,
		},
	})
}

// Wires the services of the application together and returns the adapter, which the second instance callback needs as well. Only the services that are needed to generate bindings or lifecycle hooks are registered in Wails.
func wireServices(app *application.App) *transport.Adapter {
	files := system.NewFileService()
	adapter := transport.NewAdapter(app, files)

	actions := system.Actions{Translate: adapter.Show, Open: adapter.Restore}

	app.RegisterService(application.NewService(system.NewIntegrationService(app, files, actions, trayIcon)))
	app.RegisterService(application.NewService(adapter))

	return adapter
}
