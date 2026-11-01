//go:build linux

package clipboard

import (
	"os"
	"os/exec"
	"strings"
)

const (
	sessionWayland = "wayland"
	sessionX11     = "x11"
)

// Picks the first installed tool that fits the session. Returns ErrNoTool if none is installed.
func newBackend() (backend, error) {
	return selectTool(sessionType(), exec.LookPath)
}

// Returns the first candidate with both programs installed, with their resolved paths. Returns ErrNoTool if
// there is none.
func selectTool(session string, lookup func(string) (string, error)) (backend, error) {
	for _, candidate := range candidates(session) {
		readPath, err := lookup(candidate.readBinary)
		if err != nil {
			continue
		}

		writePath, err := lookup(candidate.writeBinary)
		if err != nil {
			continue
		}

		candidate.readPath, candidate.writePath = readPath, writePath

		return candidate, nil
	}

	return nil, ErrNoTool
}

// Returns the tools in the order they are tried. An unknown session counts as Wayland; the other session's tools
// stay as a fallback, because the X11 programs also work under XWayland.
func candidates(session string) []*tool {
	if session == sessionX11 {
		return []*tool{xclipTool(), xselTool(), waylandTool()}
	}

	return []*tool{waylandTool(), xclipTool(), xselTool()}
}

// Returns the session type, falling back to the display variables when XDG_SESSION_TYPE is missing, as in a
// systemd user unit. Empty if it cannot be determined.
func sessionType() string {
	switch strings.ToLower(os.Getenv("XDG_SESSION_TYPE")) {
	case sessionWayland:
		return sessionWayland
	case sessionX11:
		return sessionX11
	}

	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return sessionWayland
	}

	if os.Getenv("DISPLAY") != "" {
		return sessionX11
	}

	return ""
}

// Returns the 'wl-clipboard' tool for Wayland. Reads without the newline wl-paste appends.
func waylandTool() *tool {
	return &tool{
		label:        "wl-clipboard",
		readBinary:   "wl-paste",
		readArgs:     []string{"--primary", "--no-newline"},
		writeBinary:  "wl-copy",
		writeArgs:    nil,
		emptyMarkers: []string{"Nothing is copied", "No selection"},
	}
}

// Returns the 'xclip' tool for X11.
func xclipTool() *tool {
	return &tool{
		label:        "xclip",
		readBinary:   "xclip",
		readArgs:     []string{"-selection", "primary", "-out"},
		writeBinary:  "xclip",
		writeArgs:    []string{"-selection", "clipboard", "-in"},
		emptyMarkers: []string{"target STRING not available"},
	}
}

// Returns the 'xsel' tool for X11. It succeeds silently on an empty selection, so it needs no marker.
func xselTool() *tool {
	return &tool{
		label:        "xsel",
		readBinary:   "xsel",
		readArgs:     []string{"--primary", "--output"},
		writeBinary:  "xsel",
		writeArgs:    []string{"--clipboard", "--input"},
		emptyMarkers: nil,
	}
}
