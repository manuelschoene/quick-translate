package clipboard

import (
	"errors"
	"sync"
)

// Returned when none of the supported clipboard programs is installed. Only Linux reads the clipboard with external programs, so this error can not occur on other operating systems.
var ErrNoTool = errors.New("Could not find a program to access the clipboard. Please install 'wl-clipboard' for a Wayland session or 'xclip' or 'xsel' for an X11 session.")

type backend interface {
	read() (string, error)
	write(text string) error
	name() string
}

type Clipboard struct {
	backend backend
	mutex   sync.Mutex
}

// Creates a new clipboard for the operating system the application runs on and picks the backend it works with. On Linux the selection is read with an external program, so ErrNoTool is returned if none of the supported programs is installed. Returns an error if the clipboard of the operating system can not be reached.
func NewClipboard() (*Clipboard, error) {
	instance, err := newBackend()
	if err != nil {
		return nil, err
	}

	return &Clipboard{backend: instance}, nil
}

// Returns the text the user has selected. On Linux this is the primary selection, which holds the text marked with the mouse, on every other operating system it is the regular clipboard the user has copied into. Returns an empty text if nothing is selected and an error if the clipboard can not be read.
func (c *Clipboard) Read() (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.backend.read()
}

// Writes the text to the regular clipboard on every operating system, so it can be pasted with Ctrl+V. The primary selection is never written to on Linux, because it belongs to the text the user has marked. Returns an error if the clipboard can not be written.
func (c *Clipboard) Write(text string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.backend.write(text)
}

// Returns the name of the backend the clipboard works with, which is the name of the package the external program belongs to on Linux and the name of the library on every other operating system. Meant for log output.
func (c *Clipboard) Backend() string {
	return c.backend.name()
}
