//go:build !linux

package system

import (
	"errors"
)

var ErrUnsupported = errors.New("Quick Translate cannot integrate with this operating system yet.")

// Removes everything the application has written for the user. Answers with ErrUnsupported as their is nothing to remove with no integration.
func Purge(_ *FileService) ([]File, error) {
	return nil, ErrUnsupported
}

// Reports what Purge would remove. Answers with ErrUnsupported for the same reason as Purge.
func Removals(_ *FileService) ([]File, error) {
	return nil, ErrUnsupported
}

// Reports what the integration into the system looks like. Answers with ErrUnsupported for the same reason as Purge.
func Status(_ *FileService) (*Report, error) {
	return nil, ErrUnsupported
}

// Only includes all entries each operating system shares.
func entries() []entry {
	return userEntries()
}

// Integrates with the part of the system the user owns. Returns no error as there is nothing to integrate.
func integrate(_ *FileService) error {
	return nil
}
