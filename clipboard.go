package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Clipboard interface {
	Read() (string, error)
	Write(string) error
}

func ClipboardInit() (Clipboard, error) {
	goos := runtime.GOOS

	if goos != "linux" {
		return nil, fmt.Errorf("Error: Unsupported operating system: %s", goos)
	}

	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType == "" {
		return nil, fmt.Errorf("Error: XDG_SESSION_TYPE environment variable is not set")
	}
	if sessionType != "wayland" {
		return nil, fmt.Errorf("Error: Unsupported session type: %s", sessionType)
	}

	return WaylandClipboard{}, nil
}

type WaylandClipboard struct{}

func (cp WaylandClipboard) Read() (string, error) {
	cmd := exec.Command("wl-paste", "-p")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error: Could not read from clipboard. Is the 'wl-clipboard' package installed?\nFailed with error: %w", err)
	}

	return string(out), nil
}

func (cp WaylandClipboard) Write(data string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(data)
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error: Could not write to clipboard. Is the 'wl-clipboard' package installed?\nFailed with error: %w", err)
	}

	return nil
}
