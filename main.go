package main

import (
	"embed"
	"log/slog"
	"os"

	"quick-translate/internal/cli"
	"quick-translate/internal/config"
	"quick-translate/internal/language"
	"quick-translate/internal/provider"
	"quick-translate/internal/store"
	"quick-translate/internal/system"
	"quick-translate/internal/translation"

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

	var open func()

	// Built here rather than inside createApp, so it closes over this variable and sees the action wireServices returns.
	app := createApp(func() {
		// Nil only between application.New() taking the lock and the wiring below.
		if open != nil {
			open()
		}
	})

	window := registerWindows(app)

	open = wireServices(app, window)

	err := app.Run()
	if err != nil {
		system.NewLogger().Error("application run", slog.Any("error", err))
		os.Exit(1)
	}
}

// Creates the application with its options, but does not yet register any services or windows. The second instance callback is handed in, because the lock it answers is taken in here, before anything it could reach exists.
func createApp(onSecondInstance func()) *application.App {
	return application.New(application.Options{
		Name:        system.AppName,
		Description: "Translate your selected text anywhere",
		Linux: application.LinuxOptions{
			ApplicationID: system.ApplicationID, // Required for application matching with desktop entry, KWin rules, etc.
		},
		Logger: system.NewLogger(),
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
func registerWindows(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            system.AppName,
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

// Wires the services of the application together and returns the action that opens the window, which the second instance callback needs as well. Only the services that are needed to generate bindings or lifecycle hooks are registered in Wails.
func wireServices(app *application.App, window *application.WebviewWindow) func() {
	files := system.NewFiles()
	notifier := system.NewNotifier(files, app.Logger)

	storage := store.New(files, notifier, app.Logger)
	cfg := config.New(files, notifier, app.Logger)
	prov := provider.New(app, cfg, app.Logger)

	// Declared before the system service, which the language service reads the locale fromsyso the actions can close over it.
	var translations *translation.Service

	show := func() { window.Show(); window.Center() }

	actions := system.Actions{
		Translate: func() { show(); translations.TranslateSelection() },
		Open: func() {
			show()

			// An empty or disabled history is no reason for an error; the window is open either way.
			if err := translations.Latest(); err != nil {
				app.Logger.Info("show latest translation", slog.Any("error", err))
			}
		},
	}

	sys := system.New(app, files, actions, trayIcon, app.Logger)
	lang := language.New(app, cfg, storage, prov, sys, app.Logger)
	translations = translation.New(app, storage, cfg, prov, lang, notifier, app.Logger)

	// Started in this order and shut down in reverse. The notifier comes first, so a failing service can still reach the user.
	app.RegisterService(application.NewService(notifier))
	app.RegisterService(application.NewService(storage))
	app.RegisterService(application.NewService(cfg))
	app.RegisterService(application.NewService(prov))
	app.RegisterService(application.NewService(lang))
	app.RegisterService(application.NewService(translations))
	app.RegisterService(application.NewService(sys))

	return actions.Open
}
