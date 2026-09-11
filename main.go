package main

import (
	"fmt"
	"os"
	"strings"

	"quick-translate/internal/desktop"
	"quick-translate/internal/transport"
)

// The commands the binary understands besides starting the application, and the flag the install command takes to be told which desktop to install for.
const (
	commandInstall   = "--install"
	commandUninstall = "--uninstall"
	commandStatus    = "--status"
	commandVersion   = "--version"
	commandHelp      = "--help"
	flagDesktop      = "--desktop="
	flagPurge        = "--purge"
)

// The text the help command prints.
// The version this binary was built as. Stamped in by the Makefile from the version in 'wails.json', or from
// the git tag when the build was made on one. A build made without the linker flag keeps the fallback, which
// is why it names itself a development build rather than a release.
var version = "0.0.0-dev"

const usage = `Quick Translate translates the text you have selected, triggered by a global shortcut.

Started without a command it runs the application, or hands the shortcut over to the instance that is
already running and stops right away.

Usage:
  quick-translate                            Start the application or trigger the running one.
  quick-translate --install [--desktop=NAME] Register this binary with the desktop and start it with the session.
  quick-translate --uninstall [--purge]      Remove the application from the desktop, but not the binary itself.
  quick-translate --status                   Show where the application has installed itself.
  quick-translate --version                  Show the version this binary was built as.
  quick-translate --help                     Show this text.

The desktop of the install command is detected from the session when it is not given. Pass 'kde' for the
Plasma installation, which binds the global shortcut and the window rules as well, or 'generic' for the one
that works on every freedesktop-compliant desktop.

The uninstall keeps your configuration, your history and the cached language lists, so installing again
picks up where you left off. Add '--purge' to delete them as well. This can not be undone.
`

// Starts the application, or hands the shortcut over to the instance that is already running and stops right away. A command given on the command line is carried out instead of any of that.
func main() {
	if handleCommand() {
		return
	}

	if transport.Connect() {
		return
	}

	// A problem that keeps the application from translating is not a reason to stop: it is shown in the window
	// the next time the shortcut is pressed, and the service manager is spared restarting a process that would
	// fail in exactly the same way for the rest of the session.
	runApp(transport.NewAdapter())
}

// Carries out the command the binary was started with and reports whether one was found, in which case the application must not start. Arguments that are none of the known commands are ignored rather than rejected, because 'wails dev' starts the binary with arguments of its own.
func handleCommand() bool {
	for _, arg := range os.Args[1:] {
		switch arg {
		case commandInstall:
			runCommand(func() error { return desktop.Install(desktopChoice()) })
		case commandUninstall:
			runCommand(uninstall)
		case commandStatus:
			runCommand(printStatus)
		case commandVersion:
			printVersion()
		case commandHelp:
			fmt.Print(usage)
		default:
			continue
		}

		return true
	}

	return false
}

// Runs a command and ends the process with a failing exit code when it reports an error, so a Makefile or a script notices an installation that did not work out.
func runCommand(command func() error) {
	if err := command(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Removes the application from the desktop and, only when it was explicitly asked for, the files it has written for the user as well. Keeping them is the default because reinstalling is far more common than leaving for good, and a history that disappears with a rebuild would be a nasty surprise.
func uninstall() error {
	if err := desktop.Uninstall(); err != nil {
		return err
	}

	if !hasFlag(flagPurge) {
		return nil
	}

	return desktop.Purge()
}

// Reports whether the given flag was given on the command line.
func hasFlag(name string) bool {
	for _, arg := range os.Args[1:] {
		if arg == name {
			return true
		}
	}

	return false
}

// Returns the desktop the install command was asked to install for, which is empty when the flag was not given and the installation detects the desktop itself.
func desktopChoice() string {
	for _, arg := range os.Args[1:] {
		if value, found := strings.CutPrefix(arg, flagDesktop); found {
			return value
		}
	}

	return ""
}

// Prints the version this binary was built as, which is what a report about a problem should name.
func printVersion() {
	fmt.Printf("Quick Translate %s\n", version)
}

// Prints where the application has installed itself and what starts it with the session. The version comes first, because a status that is pasted into a report is worth more when it says which build it describes. Returns an error if the installation can not be looked at.
func printStatus() error {
	report, err := desktop.Status()
	if err != nil {
		return err
	}

	printVersion()
	fmt.Println()
	fmt.Print(report)

	return nil
}
