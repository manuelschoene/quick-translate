package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

type x11Clipboard struct {}

func (cp x11Clipboard) Read() (string, error) {
	cmd := exec.Command("xclip", "-o")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error: Could not read from clipboard. Is xclip installed?\nError: %w", err)
	}

	return string(out), nil
}

func (cp x11Clipboard) Write(data string) error {
	cmd := exec.Command("xclip")
	cmd.Stdin = strings.NewReader(data)
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error: Could not write to clipboard. Is xclip installed?\nError: %w", err)
	}

	return nil
}