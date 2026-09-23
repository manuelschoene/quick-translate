package system

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// What the desktop session can reach in the application. Handed in, so this package stays free of the rest of it.
type Actions struct {
	Translate func() // Show the window and translate the selected text, which is what the shortcut does
	Open      func() // Show the window with the last translation, without translating anything
}

type IntegrationService struct {
	app      *application.App
	files    *FileService
	actions  Actions
	trayIcon []byte // PNG, because the tray can not read the SVG the desktop entry points at
}

// Creates a new integration service. All integration runs on startup and does not integrate with the frontend.
func NewIntegrationService(app *application.App, files *FileService, actions Actions, trayIcon []byte) *IntegrationService {
	return &IntegrationService{app: app, files: files, actions: actions, trayIcon: trayIcon}
}

// Runs the system integration based on the operating system used. The autostart registration, the global shortcut and the tray icon are ensured on every start. As failed integration is not a fatal error, every error is reported and swallowed.
func (i *IntegrationService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	if err := integrate(i.files); err != nil {
		fmt.Printf("Could not put Quick Translate into the desktop: %v\n", err)
	}

	if err := i.ensureAutostart(); err != nil {
		fmt.Printf("Could not register Quick Translate to start with the session: %v\n", err)
	}

	// Unregistering is done by Wails on shutdown. Refusal by the desktop reaches the error handler and is not returned here.
	if err := i.app.GlobalShortcut.Register(Shortcut, i.actions.Translate); err != nil {
		fmt.Printf("Could not bind the '%s' shortcut to Quick Translate: %v\n", Shortcut, err)
	}

	i.showInTray()

	return nil
}

// Registers the autostart of the application with the desktop session, so the global shortcut can reach it. Rewrittes the entry if binary was moved or the autostart entry was removed.
func (i *IntegrationService) ensureAutostart() error {
	enabled, err := i.app.Autostart.IsEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	// Application ID needed as identifier, as the session needs to match the autostart entry with the application running
	return i.app.Autostart.EnableWithOptions(application.AutostartOptions{Identifier: ApplicationID})
}

// Registers the application in the system tray, showing it as an icon in the status area. Wails starts the tray once the application runs and destroys it on shutdown. A desktop without a StatusNotifier host shows nothing and reports it through the error handler of the application.
func (i *IntegrationService) showInTray() {
	menu := application.NewMenu()

	menu.Add("Open").OnClick(func(_ *application.Context) { i.actions.Open() })
	menu.Add("Translate selection").OnClick(func(_ *application.Context) { i.actions.Translate() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(_ *application.Context) { i.app.Quit() })

	tray := i.app.SystemTray.New()

	tray.SetLabel(TrayLabel)
	tray.SetTooltip(TrayTooltip)
	tray.SetIcon(i.trayIcon)
	tray.SetMenu(menu)

	// Set explicitly, because Wails only falls back to a default click handler for a tray an application window is attached to.
	tray.OnClick(i.actions.Open)
}
