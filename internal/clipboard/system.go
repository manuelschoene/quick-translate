//go:build !linux

package clipboard

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	system "golang.design/x/clipboard"
)

type systemClipboard struct{}

// Creates the backend for every operating system besides Linux. Initializes the library, so an unreachable
// clipboard fails the startup and not the first translation.
func newBackend() (backend, error) {
	if err := system.Init(); err != nil {
		return nil, fmt.Errorf("clipboard: initialize on %s, is a graphical session running: %w", runtime.GOOS, err)
	}

	return systemClipboard{}, nil
}

// Returns the text of the regular clipboard, empty if it holds no text.
func (c systemClipboard) read(ctx context.Context) (string, error) {
	text, err := system.Read(ctx, system.FmtText)
	if errors.Is(err, system.ErrNoData) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("clipboard: read: %w", err)
	}

	return string(text), nil
}

// Writes the text to the regular clipboard. The returned channel only reports a later takeover and is ignored.
func (c systemClipboard) write(ctx context.Context, text string) error {
	if _, err := system.Write(ctx, system.FmtText, []byte(text)); err != nil {
		return fmt.Errorf("clipboard: write: %w", err)
	}

	return nil
}

// Returns the library name.
func (c systemClipboard) name() string {
	return "golang.design/x/clipboard"
}
