package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"quick-translate/internal/core"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Sent when the window was opened by the shortcut and the translation is on its way, so the frontend can show that it is working instead of the previous translation.
const eventTranslating = "translating"

// Sent with the TranslationDto of a translation that was started by the shortcut.
const eventTranslation = "translation"

// Sent with the message of a translation that was started by the shortcut and failed.
const eventError = "error"

type Adapter struct {
	ctx      context.Context
	core     *core.Core
	listener net.Listener
	mutex    sync.Mutex
}

// Creates a new adapter together with the core it works on. Returns an error for every problem that stops the application from running, so an adapter that is returned is ready to serve the frontend.
func NewAdapter() (*Adapter, error) {
	c, err := core.NewCore()
	if err != nil {
		return nil, err
	}

	return &Adapter{core: c}, nil
}

// Prepares the application once the window is ready and starts listening for the shortcut. Called by Wails, which also hands over the context every call into its runtime needs.
func (a *Adapter) StartUp(ctx context.Context) {
	a.ctx = ctx

	if err := a.listenOnSocket(); err != nil {
		// The application is started hidden and is only brought up through the socket, so it is of no use without it.
		fmt.Println(err)
		runtime.Quit(ctx)
	}
}

// Releases everything the application holds. Called by Wails on shutdown, so the history database is closed and the socket is removed even when the application is stopped from the outside.
func (a *Adapter) Shutdown(ctx context.Context) {
	// Closing the listener also removes the socket file, because it was created by this instance.
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			fmt.Printf("Could not stop listening for the shortcut: %v\n", err)
		}
	}

	a.core.Close()
}

// Hides the window without stopping the application, which is what the frontend does when the user dismisses it.
func (a *Adapter) Hide() {
	runtime.WindowHide(a.ctx)
}

// Shows the window and translates the selected text. Used for the shortcut, which reaches the application through the socket instead of the frontend, so the result is sent as an event. The window is shown before the translation is made, so it reacts to the shortcut right away.
func (a *Adapter) show() {
	runtime.WindowShow(a.ctx)
	runtime.WindowCenter(a.ctx)
	runtime.EventsEmit(a.ctx, eventTranslating)

	dto, err := a.translateFromClipboard()
	if err != nil {
		fmt.Printf("Could not translate the selected text: %v\n", err)
		runtime.EventsEmit(a.ctx, eventError, err.Error())
		return
	}

	runtime.EventsEmit(a.ctx, eventTranslation, dto)
}

// Translates the selected text and returns the new state of the translation view.
func (a *Adapter) translateFromClipboard() (*TranslationDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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
