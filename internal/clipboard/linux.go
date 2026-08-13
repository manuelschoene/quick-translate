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

// Creates the backend for Linux by picking the first installed program that fits the current session. Returns ErrNoTool if none of the supported programs is installed.
func newBackend() (backend, error) {
	return selectTool(sessionType(), exec.LookPath)
}

// Returns the first of the candidates whose programs are both installed. The lookup is handed in so the selection can be checked without the programs being installed. Returns ErrNoTool if none of the candidates is installed.
func selectTool(session string, lookup func(string) (string, error)) (backend, error) {
	for _, candidate := range candidates(session) {
		if _, err := lookup(candidate.readBinary); err != nil {
			continue
		}

		if _, err := lookup(candidate.writeBinary); err != nil {
			continue
		}

		return candidate, nil
	}

	return nil, ErrNoTool
}

// Returns the supported tools in the order they should be tried for the given session type. An unknown session type is treated like a Wayland session, and the tools of the other session type stay in the list as a fallback, because the X11 programs also work under XWayland.
func candidates(session string) []*tool {
	if session == sessionX11 {
		return []*tool{xclipTool(), xselTool(), waylandTool()}
	}

	return []*tool{waylandTool(), xclipTool(), xselTool()}
}

// Returns the type of the graphical session the application runs in. The session variable is missing when the application is started as a systemd user unit that does not know the graphical environment, so the display variables are used as a fallback. Returns an empty string if the session type can not be determined.
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

// Returns the tool for the 'wl-clipboard' package, which is used in Wayland sessions. The selection is read without the newline the program appends on its own.
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

// Returns the tool for the 'xclip' package, which is used in X11 sessions.
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

// Returns the tool for the 'xsel' package, which is used in X11 sessions. The program stays silent and succeeds on an empty selection, so it does not need a marker for it.
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
