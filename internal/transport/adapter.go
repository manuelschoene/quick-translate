package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"quick-translate/internal/core"
	"quick-translate/internal/system"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Sent when the window was opened by the shortcut and the translation is on its way, so the frontend can show that it is working instead of the previous translation.
const eventTranslating = "translating"

// Sent with the TranslationDto of a translation that was started by the shortcut.
const eventTranslation = "translation"

// Sent with the message of a translation that was started by the shortcut and failed.
const eventError = "error"

type Adapter struct {
	app      *application.App
	files    *system.FileService
	core     *core.Core
	listener net.Listener
	mutex    sync.Mutex
}

// Creates a new adapter and builds the core it works on. A core that can not be built does not stop the application: the reason is almost always a configuration the user has yet to finish, and a window that says so is worth more than a process that exits and is restarted by the service manager for as long as the session lasts. The adapter answers every call with that reason instead, and tries again on the next one.
func NewAdapter(app *application.App, files *system.FileService) *Adapter {
	application.RegisterEvent[application.Void](eventTranslating)
	application.RegisterEvent[*TranslationDto](eventTranslation)
	application.RegisterEvent[string](eventError)

	adapter := &Adapter{
		app:   app,
		files: files,
	}

	// Built here rather than on the first call, so a problem shows up in the log when the session starts and the languages of the provider are fetched before the user asks for the first translation.
	if err := adapter.ready(); err != nil {
		fmt.Printf("Quick Translate can not translate yet: %v\n", err)
	}

	return adapter
}

// Builds the core if it is not standing yet and reports what keeps it from being built. Every call from the frontend and every shortcut goes through this, so a core that was never built is answered with the reason instead of reaching a nil pointer. Building it again on every call is what lets a configuration that was corrected while the application was running take effect with the next shortcut, rather than asking the user to restart the service. Requires the lock to be held.
func (a *Adapter) ready() error {
	if a.core != nil {
		return nil
	}

	instance, err := core.NewCore(a.files)
	if err != nil {
		return err
	}

	a.core = instance

	return nil
}

// Called by Wails on startup. Starts listening on the IPC socket for the shortcut. Returns an error if the socket cannot be listened on, therefore shutting down the application as it would not work otherwise.
func (a *Adapter) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	if err := a.listenOnSocket(); err != nil {
		return err
	}

	return nil
}

// Called by Wails on shutdown. Close socket and database connection.
func (a *Adapter) ServiceShutdown() error {
	// Closing the listener also removes the socket file, because it was created by this instance.
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			fmt.Printf("Could not stop listening for the shortcut: %v\n", err)
		}
	}

	// The core is missing when it could never be built, which is the state the adapter serves the reason for.
	if a.core != nil {
		a.core.Close()
	}

	return nil
}

// Hides the window without stopping the application, which is what the frontend does when the user dismisses it.
func (a *Adapter) Hide() {
	a.app.Window.Current().Hide()
}

// Shows the window and translates the selected text. Used for the shortcut, which reaches the application through the socket instead of the frontend, so the result is sent as an event. The window is shown before the translation is made, so it reacts to the shortcut right away.
func (a *Adapter) show() {
	a.app.Window.Current().Show()
	a.app.Window.Current().Center()
	a.app.Event.Emit(eventTranslating)

	dto, err := a.translateFromClipboard()
	if err != nil {
		fmt.Printf("Could not translate the selected text: %v\n", err)
		a.app.Event.Emit(eventError, err.Error())
		return
	}

	a.app.Event.Emit(eventTranslation, dto)
}

// Translates the selected text and returns the new state of the translation view.
func (a *Adapter) translateFromClipboard() (*TranslationDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.ready(); err != nil {
		return nil, err
	}

	if _, err := a.core.TranslateFromClipboard(); err != nil {
		return nil, err
	}

	return a.translation(), nil
}

// Translates the text of the current translation again, which is done after the languages or the provider were changed. A missing translation and missing languages are not reported as an error, because both are states the user passes through before the first translation is made and both are visible in the returned state anyway. Requires the lock to be held.
func (a *Adapter) retranslate() error {
	_, err := a.core.TranslateFromHistory()

	switch {
	case err == nil:
		return nil
	case errors.Is(err, core.ErrNoTranslation),
		errors.Is(err, core.ErrNoSourceLanguage),
		errors.Is(err, core.ErrNoTargetLanguage):
		return nil
	default:
		return err
	}
}
