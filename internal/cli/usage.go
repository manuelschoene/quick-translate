package cli

import (
	"fmt"
	"strings"
)

const (
	preamble = `Quick Translate translates your selected text, triggered by a global shortcut.

Usage: quick-translate <command>
When no command is given, the application is started or the running one is triggered.

Available commands:`

	epilogue = "Report bugs at https://github.com/manuelschoene/quick-translate/issues"
)

// Prints the summary of each command as one unified usage of the binary. Never fails.
func printUsage(out interaction, _ string, _ []string) error {
	var text strings.Builder

	text.WriteString(preamble)
	text.WriteString("\n")
	text.WriteString("\n")

	for _, current := range commands() {
		fmt.Fprintf(&text, "  %-10s %s\n", current.flag, current.summary)
	}

	text.WriteString("\n")
	text.WriteString(epilogue)
	text.WriteString("\n")
	fmt.Fprint(out, text.String())

	return nil
}

// Prints the version this binary was built as. Never fails.
func printVersion(out interaction, version string, _ []string) error {
	fmt.Fprintf(out, "Quick Translate %s\n", version)

	return nil
}
