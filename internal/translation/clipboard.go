package translation

import (
	"context"
	"errors"
	"fmt"

	"golang.design/x/clipboard"
)

// Connects to the clipboard once. Fails if there is none, e.g. without a graphical session.
func startClipboard() error {
	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("translation: initialize clipboard, is a graphical session running: %w", err)
	}

	return nil
}

// Returns the selected text, empty if nothing is selected. Reads the primary selection (whatever is marked with
// the mouse) and falls back to the regular clipboard where there is none: outside X11 and Wayland, or under a
// compositor without primary selection support.
func readSelection(ctx context.Context) (string, error) {
	text, err := clipboard.Read(ctx, clipboard.FmtText, clipboard.FromPrimary())
	if errors.Is(err, clipboard.ErrUnsupported) {
		text, err = clipboard.Read(ctx, clipboard.FmtText)
	}

	if errors.Is(err, clipboard.ErrNoData) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("translation: read selection: %w", err)
	}

	return string(text), nil
}

// Writes the text to the regular clipboard, never to the primary selection. The running application serves it,
// so the returned takeover channel is ignored.
func writeClipboard(ctx context.Context, text string) error {
	if _, err := clipboard.Write(ctx, clipboard.FmtText, []byte(text)); err != nil {
		return fmt.Errorf("translation: write clipboard: %w", err)
	}

	return nil
}
