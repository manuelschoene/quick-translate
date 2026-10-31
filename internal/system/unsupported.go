//go:build !linux

package system

import (
	"errors"
)

var ErrUnsupported = errors.New("system: operating system not supported")

// Removes everything the application has written for the user. Answers with ErrUnsupported as their is nothing to remove with no integration.
func (f *Files) Purge() ([]File, error) {
	return nil, ErrUnsupported
}

// Reports what Purge would remove. Answers with ErrUnsupported for the same reason as Purge.
func (f *Files) Removals() ([]File, error) {
	return nil, ErrUnsupported
}

// Reports what the integration into the system looks like. Answers with ErrUnsupported for the same reason as Purge.
func (f *Files) Status() (*Report, error) {
	return nil, ErrUnsupported
}

// Only includes all entries each operating system shares.
func entries() []entry {
	return userEntries()
}

// Integrates with the part of the system the user owns. Returns no error as there is nothing to integrate.
func integrate(_ *Files) error {
	return nil
}

// No configuration needed for non-Linux systems, as the default logger is already suitable.
func configureLogger() {}

// Answers with ErrUnsupported, as the locale is not detected on this operating system yet.
func locale() (string, error) {
	return "", ErrUnsupported
}
