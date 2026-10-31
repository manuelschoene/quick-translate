package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	preamble = `Quick Translate translates your selected text, triggered by a global shortcut.

Usage: quick-translate [-verbose | -silent] <command> [flags]
When no command is given, the application is started, or the running one is brought back up.`

	epilogue = "Report bugs at https://github.com/manuelschoene/quick-translate/issues"
)

// Prints the usage of the whole binary, or of the one command named by the argument. Fails for an unknown command.
func printHelp(out interaction, _ string, args []string) error {
	if len(args) == 0 {
		printUsage(out)

		return nil
	}

	found := findCommand(args[0])
	if found == nil {
		return usageError{fmt.Errorf("unknown command %q", args[0])}
	}

	printCommandUsage(out, found)

	return nil
}

// Prints the summary of each command and the global flags as one unified usage of the binary.
func printUsage(out io.Writer) {
	var text strings.Builder

	text.WriteString(preamble)
	text.WriteString("\n\nAvailable commands:\n\n")

	for _, current := range commands() {
		fmt.Fprintf(&text, "  %-16s %s\n", strings.TrimSpace(current.name+" "+current.args), current.summary)
	}

	root, _, _ := globalFlags()

	text.WriteString("\nGlobal flags:\n\n")
	writeFlags(&text, root)

	text.WriteString("\nRun 'quick-translate help <command>' to see the flags of a command.\n\n")
	text.WriteString(epilogue)
	text.WriteString("\n")
	fmt.Fprint(out, text.String())
}

// Prints the usage line, the summary and the flags of one command.
func printCommandUsage(out io.Writer, c *command) {
	fs, _ := c.flags()

	var text strings.Builder

	fmt.Fprintf(&text, "Usage: quick-translate [-verbose | -silent] %s", c.name)

	if c.confirms {
		text.WriteString(" [-yes]")
	}

	if len(c.args) != 0 {
		fmt.Fprintf(&text, " %s", c.args)
	}

	fmt.Fprintf(&text, "\n\n%s\n", c.summary)

	if c.confirms {
		text.WriteString("\nFlags:\n\n")
		writeFlags(&text, fs)
	}

	fmt.Fprint(out, text.String())
}

// Writes every flag of the set with its description, one per line, in the order of their names.
func writeFlags(text *strings.Builder, fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(text, "  %-16s %s\n", "-"+f.Name, f.Usage)
	})
}

// Prints the version this binary was built as. Never fails.
func printVersion(out interaction, version string, _ []string) error {
	fmt.Fprintf(out, "Quick Translate %s\n", version)

	return nil
}
