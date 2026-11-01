package system

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// What the desktop session can reach in the application. Handed in, so this package stays free of the rest of it.
type Actions struct {
	Translate func() // Show the window and translate the selected text, which is what the shortcut does
	Open      func() // Show the window with the last translation, without translating anything
}

type Service struct {
	app      *application.App
	files    *Files
	actions  Actions
	trayIcon []byte // PNG, because the tray can not read the SVG the desktop entry points at
	logger   *slog.Logger
}

// Creates the service. All integration runs on startup. Does not integrate with the frontend.
func New(app *application.App, files *Files, actions Actions, trayIcon []byte, logger *slog.Logger) *Service {
	return &Service{
		app:      app,
		files:    files,
		actions:  actions,
		trayIcon: trayIcon,
		logger:   logger.With("category", "system"),
	}
}

// Runs the system integration based on the operating system used. The autostart registration, the global shortcut and the tray icon are ensured on every start. As failed integration is not a fatal error, every error is logged and swallowed.
func (s *Service) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	if err := integrate(s.files); err != nil {
		s.logger.Warn("integrate into the desktop", slog.Any("error", err))
	}

	if err := s.ensureAutostart(); err != nil {
		s.logger.Warn("register autostart", slog.Any("error", err))
	}

	// Unregistering is done by Wails on shutdown. Refusal by the desktop reaches the error handler and is not returned here.
	if err := s.app.GlobalShortcut.Register(Shortcut, s.actions.Translate); err != nil {
		s.logger.Warn("bind global shortcut", slog.String("shortcut", Shortcut), slog.Any("error", err))
	}

	s.showInTray()

	return nil
}

// Returns the user's locale as a BCP 47 tag, e.g. "de-DE". Fails if the operating system does not reveal one.
//
//wails:ignore
func (s *Service) Locale() (string, error) {
	return locale()
}

// Registers the autostart of the application with the desktop session, so the global shortcut can reach it. Rewrittes the entry if binary was moved or the autostart entry was removed.
func (s *Service) ensureAutostart() error {
	enabled, err := s.app.Autostart.IsEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	// Application ID needed as identifier, as the session needs to match the autostart entry with the application running
	return s.app.Autostart.EnableWithOptions(application.AutostartOptions{Identifier: ApplicationID})
}

// Registers the application in the system tray, showing it as an icon in the status area. Wails starts the tray once the application runs and destroys it on shutdown. A desktop without a StatusNotifier host shows nothing and reports it through the error handler of the application.
func (s *Service) showInTray() {
	menu := application.NewMenu()

	menu.Add("Open").OnClick(func(_ *application.Context) { s.actions.Open() })
	menu.Add("Translate selection").OnClick(func(_ *application.Context) { s.actions.Translate() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(_ *application.Context) { s.app.Quit() })

	tray := s.app.SystemTray.New()

	tray.SetLabel(AppName)
	tray.SetTooltip(trayTooltip)
	tray.SetIcon(s.trayIcon)
	tray.SetMenu(menu)

	// Set explicitly, because Wails only falls back to a default click handler for a tray an application window is attached to.
	tray.OnClick(s.actions.Open)
}
