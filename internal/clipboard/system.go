//go:build !linux

package clipboard

import (
	"fmt"
	"runtime"

	system "golang.design/x/clipboard"
)

type systemClipboard struct{}

// Creates the backend for every operating system besides Linux, which reads and writes the regular clipboard through the library. The library is initialized here, so an unreachable clipboard is reported when the application starts and not when the user asks for the first translation. Returns an error if the clipboard of the operating system can not be reached.
func newBackend() (backend, error) {
	if err := system.Init(); err != nil {
		return nil, fmt.Errorf("Could not initialize the clipboard of %s: %w. Please make sure the application is started inside a graphical session.", runtime.GOOS, err)
	}

	return systemClipboard{}, nil
}

// Returns the text of the regular clipboard, which the user has copied with Ctrl+C. The library reports an unreadable clipboard the same way as an empty one, so an empty text is returned in both cases.
func (c systemClipboard) read() (string, error) {
	return string(system.Read(system.FmtText)), nil
}

// Writes the text to the regular clipboard. The library answers with a channel that reports when another application takes the clipboard over, which is of no interest here and must not be waited for. A missing channel is the only way the library signals a failed write.
func (c systemClipboard) write(text string) error {
	if system.Write(system.FmtText, []byte(text)) == nil {
		return fmt.Errorf("Could not write to the clipboard of %s. Please make sure no other application is blocking the clipboard.", runtime.GOOS)
	}

	return nil
}

// Returns the name of the library the backend works with.
func (c systemClipboard) name() string {
	return "golang.design/x/clipboard"
}
