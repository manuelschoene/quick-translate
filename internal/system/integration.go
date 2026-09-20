package system

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type IntegrationService struct {
	app      *application.App
	files    *FileService
	shortcut func() // What the global shortcut triggers. Handed in, so this package stays free of the rest of the application.
}

// Creates a new integration service. All integration runs on startup and does not integrate with the frontend.
func NewIntegrationService(app *application.App, files *FileService, shortcut func()) *IntegrationService {
	return &IntegrationService{app: app, files: files, shortcut: shortcut}
}

// Runs the system integration based on the operating system used. The autostart registration and the global shortcut are ensured on every start. As failed integration is not a fatal error, every error is reported and swallowed.
func (i *IntegrationService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	if err := integrate(i.files); err != nil {
		fmt.Printf("Could not put Quick Translate into the desktop: %v\n", err)
	}

	if err := i.ensureAutostart(); err != nil {
		fmt.Printf("Could not register Quick Translate to start with the session: %v\n", err)
	}

	// Unregistering is done by Wails on shutdown. Refusal by the desktop reaches the error handler and is not returned here.
	if err := i.app.GlobalShortcut.Register(Shortcut, i.shortcut); err != nil {
		fmt.Printf("Could not bind the '%s' shortcut to Quick Translate: %v\n", Shortcut, err)
	}

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
