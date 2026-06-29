package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

type waylandClipboard struct{}

func (cp waylandClipboard) Read() (string, error) {
	cmd := exec.Command("wl-paste", "-p")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error: Could not read from clipboard. Is the 'wl-clipboard' package installed?\nFailed with error: %w", err)
	}

	return string(out), nil
}

func (cp waylandClipboard) Write(data string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(data)
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error: Could not write to clipboard. Is the 'wl-clipboard' package installed?\nFailed with error: %w", err)
	}

	return nil
}
