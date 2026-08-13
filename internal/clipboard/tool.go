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

// The time an external program gets to answer before it is killed.
const commandTimeout = 2 * time.Second

type tool struct {
	label        string   // Name of the package the programs belong to.
	readBinary   string   // Program that prints the primary selection.
	readArgs     []string // Arguments that make the program print the primary selection unchanged.
	writeBinary  string   // Program that takes the text on its standard input.
	writeArgs    []string // Arguments that make the program fill the regular clipboard.
	emptyMarkers []string // Messages the read program prints when nothing is selected.
}

// Returns the name of the package the programs of this tool belong to.
func (t *tool) name() string {
	return t.label
}

// Returns the content of the primary selection by running the read program. The read programs print the selection and exit without leaving a process behind, so their output is collected in buffers. Returns an empty text if nothing is selected and an error if the program fails or does not answer in time.
func (t *tool) read() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, t.readBinary, t.readArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// A program that leaves a process behind would keep the pipes of the buffers open and hold up the call even after the timeout has killed it, so the pipes are released once the timeout has passed.
	cmd.WaitDelay = commandTimeout

	err := cmd.Run()
	if err == nil {
		return stdout.String(), nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("Could not read the primary selection because '%s' did not answer within %s. Please make sure the display server is running and reachable.", t.readBinary, commandTimeout)
	}

	// An empty selection is reported as a failure by some of the programs, which is the normal state and not an error.
	if stdout.Len() == 0 && t.reportsEmpty(stderr.String()) {
		return "", nil
	}

	return "", fmt.Errorf("Could not read the primary selection with '%s': %w. The program reported: %s", t.readBinary, err, message(stderr.String()))
}

// Writes the text to the regular clipboard by running the write program. Returns an error if the program does not take the text or does not start in time.
func (t *tool) write(text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.writeBinary, t.writeArgs...)
	cmd.Stdin = strings.NewReader(text)

	// The write programs fork into the background to serve the clipboard until another application takes it over. A buffer for the output would be filled through a pipe that the forked program inherits, which would keep the call waiting until the clipboard is taken over, so the output is dropped and the error output is passed on as a file. The wait delay releases the pipe of the standard input in case a program forks before it has read the whole text.
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	cmd.WaitDelay = commandTimeout

	err := cmd.Run()
	if err == nil || errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}

	return fmt.Errorf("Could not write to the clipboard with '%s': %w. Please make sure the display server is running and reachable.", t.writeBinary, err)
}

// Checks if the error output of the read program says that nothing is selected.
func (t *tool) reportsEmpty(output string) bool {
	for _, marker := range t.emptyMarkers {
		if strings.Contains(output, marker) {
			return true
		}
	}

	return false
}

// Returns the last line the program has written to its error output, or a placeholder if it stayed silent.
func message(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	last := strings.TrimSpace(lines[len(lines)-1])
	if last == "" {
		return "nothing"
	}

	return last
}
