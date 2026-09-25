package cli

import (
	"fmt"
	"os"
)

type command struct {
	flag    string                                                     // Unique flag including the leading dashes
	summary string                                                     // Description shown in the usage text
	run     func(out interaction, version string, args []string) error // args are the arguments that followed the command itself
}

const (
	codeOk      = 0
	codeFailed  = 1
	codeUnknown = 2
)

// Lists all commands available through the CLI, in order shown by the usage text. Register another command here to be reachable through the CLI.
func commands() []command {
	return []command{
		{
			flag:    "--help",
			summary: "Show this text.",
			run:     printUsage,
		},
		{
			flag:    "--purge",
			summary: "Remove everything the application has written. This can not be undone.",
			run:     purge,
		},
		{
			flag:    "--status",
			summary: "Show how the application is integrated into your desktop.",
			run:     printStatus,
		},
		{
			flag:    "--version",
			summary: "Show the version this binary was built as.",
			run:     printVersion,
		},
	}
}

// Carries out the command the binary was started with. Reports whether a command was run and the exit code a run command should exit with. The first argument has to be a command, or the call is refused. Additional arguments are passed to the command, which may or may not use them.
func Run(args []string, version string) (bool, int) {
	if len(args) == 0 {
		return false, codeOk
	}

	found := findCommand(args[0])
	if found == nil {
		fmt.Fprintf(os.Stderr, "Command not found: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Try 'quick-translate --help' to see all available commands.")

		return true, codeUnknown
	}

	// One instance for the whole run, so input buffering can take effect
	out := newTerminal(os.Stdin, os.Stdout)

	if err := found.run(out, version, args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return true, codeFailed
	}

	return true, codeOk
}

// Converts the argument into a command, or nil if no command matches.
func findCommand(arg string) *command {
	table := commands()

	for i := range table {
		if table[i].flag == arg {
			return &table[i]
		}
	}

	return nil
}
