package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"quick-translate/internal/system"
)

type command struct {
	name      string                                                     // Unique subcommand name
	args      string                                                     // Positional arguments shown in the usage text, e.g. "[command]"
	summary   string                                                     // Description shown in the usage text
	arguments int                                                        // Maximum number of positional arguments, more are refused
	confirms  bool                                                       // Asks the user, so the command gets the -yes flag
	run       func(out interaction, version string, args []string) error // args are the positional arguments after the flags
}

// How much the CLI writes, set by the global flags.
type verbosity int

const (
	normal verbosity = iota
	verbose
	silent
)

const (
	codeOk      = 0
	codeFailed  = 1
	codeUnknown = 2
)

const hint = "Try 'quick-translate help' to see all available commands."

// Returned by a command for a command line it cannot run, so Run exits with codeUnknown and the hint instead of codeFailed. The message is the one of the wrapped error.
type usageError struct{ error }

// Lists all commands available through the CLI, in order shown by the usage text. Register another command here to be reachable through the CLI.
func commands() []command {
	return []command{
		{
			name:      "help",
			args:      "[command]",
			summary:   "Show this text, or the help of one command.",
			arguments: 1,
			run:       printHelp,
		},
		{
			name:     "purge",
			summary:  "Remove everything the application has written. This can not be undone.",
			confirms: true,
			run:      purge,
		},
		{
			name:    "status",
			summary: "Show how the application is integrated into your desktop.",
			run:     printStatus,
		},
		{
			name:    "version",
			summary: "Show the version this binary was built as.",
			run:     printVersion,
		},
	}
}

// Carries out the command the binary was started with. Reports whether a command was run and the exit code a run command should exit with. Without any argument, nothing is run and the application starts. The global flags have to come before the command, its own flags after it.
func Run(args []string, version string) (bool, int) {
	if len(args) == 0 {
		return false, codeOk
	}

	root, verboseFlag, silentFlag := globalFlags()

	if code, ok := parse(root, args, printUsage); !ok {
		return true, code
	}

	if *verboseFlag && *silentFlag {
		return true, refuse(errors.New("-verbose and -silent cannot be combined"))
	}

	if root.NArg() == 0 {
		return true, refuse(errors.New("no command given"))
	}

	found := findCommand(root.Arg(0))
	if found == nil {
		return true, refuse(fmt.Errorf("unknown command %q", root.Arg(0)))
	}

	fs, yes := found.flags()

	if code, ok := parse(fs, root.Args()[1:], func(out io.Writer) { printCommandUsage(out, found) }); !ok {
		return true, code
	}

	if fs.NArg() > found.arguments {
		return true, refuse(fmt.Errorf("too many arguments for %s: %q", found.name, fs.Args()))
	}

	level := levelOf(*verboseFlag, *silentFlag)
	slog.SetLogLoggerLevel(logLevel(level))

	// One instance for the whole run, so input buffering can take effect
	out := newTerminal(os.Stdin, os.Stdout, level, *yes)

	if err := found.run(out, version, fs.Args()); err != nil {
		if errors.As(err, new(usageError)) {
			return true, refuse(err)
		}

		fmt.Fprintln(os.Stderr, err)

		return true, codeFailed
	}

	return true, codeOk
}

// Creates the flag set of the global flags, which come before the command, and returns the values they are parsed into.
func globalFlags() (*flag.FlagSet, *bool, *bool) {
	fs := flag.NewFlagSet(system.AppSlug, flag.ContinueOnError)
	verboseFlag := fs.Bool("verbose", false, "Show what the application does in detail.")
	silentFlag := fs.Bool("silent", false, "Only show errors and questions.")

	return fs, verboseFlag, silentFlag
}

// Creates the flag set of the command and returns the value of -yes, which is false for a command that never asks.
func (c *command) flags() (*flag.FlagSet, *bool) {
	fs := flag.NewFlagSet(c.name, flag.ContinueOnError)

	yes := new(bool)
	if c.confirms {
		yes = fs.Bool("yes", false, "Answer every question with yes.")
	}

	return fs, yes
}

// Parses the arguments into the flag set without letting it print anything. Asking for help prints the given usage and succeeds, any other failure is refused with a hint. Reports the exit code and whether the caller may continue.
func parse(fs *flag.FlagSet, args []string, usage func(io.Writer)) (int, bool) {
	fs.SetOutput(io.Discard)

	err := fs.Parse(args)
	switch {
	case err == nil:
		return codeOk, true
	case errors.Is(err, flag.ErrHelp):
		usage(os.Stdout)

		return codeOk, false
	default:
		return refuse(err), false
	}
}

// Prints a usage error with the hint to the help and returns the exit code for it.
func refuse(err error) int {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintln(os.Stderr, hint)

	return codeUnknown
}

// Converts the global flags into a verbosity. Both flags set at once must be refused before.
func levelOf(verboseFlag bool, silentFlag bool) verbosity {
	switch {
	case verboseFlag:
		return verbose
	case silentFlag:
		return silent
	default:
		return normal
	}
}

// Returns the lowest level logged at the verbosity. The terminal only shows warnings by default, as the info logs are meant for the journal of the running application.
func logLevel(level verbosity) slog.Level {
	switch level {
	case verbose:
		return slog.LevelDebug
	case silent:
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// Converts the argument into a command, or nil if no command matches.
func findCommand(name string) *command {
	table := commands()

	for i := range table {
		if table[i].name == name {
			return &table[i]
		}
	}

	return nil
}
