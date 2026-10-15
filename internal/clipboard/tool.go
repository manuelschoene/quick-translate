//go:build linux

package clipboard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// How long a program gets to answer before it is killed.
const commandTimeout = 2 * time.Second

type tool struct {
	label        string // Package the programs belong to
	readBinary   string // Prints the primary selection
	readArgs     []string
	writeBinary  string // Fills the regular clipboard from stdin
	writeArgs    []string
	emptyMarkers []string // What the read program prints when nothing is selected
	readPath     string   // Resolved at selection, so a later PATH cannot swap the program
	writePath    string
}

// Returns the package name.
func (t *tool) name() string {
	return t.label
}

// Returns the primary selection by running the read program, empty if nothing is selected. Fails if the program
// fails, does not answer in time or ctx is done.
func (t *tool) read(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("clipboard: read: %w", err)
	}

	timeout, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(timeout, t.readPath, t.readArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// A program that leaves a process behind would keep the pipes open past the kill, so they are released after
	// the wait delay.
	cmd.WaitDelay = commandTimeout

	err := cmd.Run()
	if err == nil {
		return stdout.String(), nil
	}

	if ctx.Err() != nil {
		return "", fmt.Errorf("clipboard: read: %w", ctx.Err())
	}

	if errors.Is(timeout.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("clipboard: '%s' did not answer within %s, is the display server reachable", t.readBinary, commandTimeout)
	}

	// Some programs fail on an empty selection, which is the normal state.
	if stdout.Len() == 0 && t.reportsEmpty(stderr.String()) {
		return "", nil
	}

	return "", fmt.Errorf("clipboard: read with '%s' (%s): %w", t.readBinary, message(stderr.String()), err)
}

// Writes the text to the regular clipboard by running the write program. Fails if the program does not take the
// text, does not start in time or ctx is done.
func (t *tool) write(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("clipboard: write: %w", err)
	}

	timeout, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeout, t.writePath, t.writeArgs...)
	cmd.Stdin = strings.NewReader(text)

	// The write programs fork into the background to serve the clipboard. A pipe for the output would be inherited
	// by the fork and keep the call waiting until the clipboard is taken over, so the output is dropped and the
	// error output passed on as a file. The wait delay releases stdin if a program forks before reading all of it.
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	cmd.WaitDelay = commandTimeout

	err := cmd.Run()
	if err == nil || errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}

	if ctx.Err() != nil {
		return fmt.Errorf("clipboard: write: %w", ctx.Err())
	}

	return fmt.Errorf("clipboard: write with '%s', is the display server reachable: %w", t.writeBinary, err)
}

// Reports whether the error output says nothing is selected.
func (t *tool) reportsEmpty(output string) bool {
	for _, marker := range t.emptyMarkers {
		if strings.Contains(output, marker) {
			return true
		}
	}

	return false
}

// Returns the last line of the error output, or a placeholder if it is empty.
func message(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	last := strings.TrimSpace(lines[len(lines)-1])
	if last == "" {
		return "no output"
	}

	return last
}
