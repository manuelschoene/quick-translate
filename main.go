package main

import (
	"fmt"
	"log"
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
	commandHelp      = "--help"
	flagDesktop      = "--desktop="
)

// The text the help command prints.
const usage = `Quick Translate translates the text you have selected, triggered by a global shortcut.

Started without a command it runs the application, or hands the shortcut over to the instance that is
already running and stops right away.

Usage:
  quick-translate                            Start the application or trigger the running one.
  quick-translate --install [--desktop=NAME] Register this binary with the desktop and start it with the session.
  quick-translate --uninstall                Remove the application from the desktop, but not the binary itself.
  quick-translate --status                   Show where the application has installed itself.
  quick-translate --help                     Show this text.

The desktop of the install command is detected from the session when it is not given. Pass 'kde' for the
Plasma installation, which binds the global shortcut and the window rules as well, or 'generic' for the one
that works on every freedesktop-compliant desktop.
`

// Starts the application, or hands the shortcut over to the instance that is already running and stops right away. A command given on the command line is carried out instead of any of that.
func main() {
	if handleCommand() {
		return
	}

	if transport.Connect() {
		return
	}

	adapter, err := transport.NewAdapter()
	if err != nil {
		log.Fatalf("Could not start Quick Translate: %v", err)
	}

	runApp(adapter)
}

// Carries out the command the binary was started with and reports whether one was found, in which case the application must not start. Arguments that are none of the known commands are ignored rather than rejected, because 'wails dev' starts the binary with arguments of its own.
func handleCommand() bool {
	for _, arg := range os.Args[1:] {
		switch arg {
		case commandInstall:
			runCommand(func() error { return desktop.Install(desktopChoice()) })
		case commandUninstall:
			runCommand(desktop.Uninstall)
		case commandStatus:
			runCommand(printStatus)
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

// Returns the desktop the install command was asked to install for, which is empty when the flag was not given and the installation detects the desktop itself.
func desktopChoice() string {
	for _, arg := range os.Args[1:] {
		if value, found := strings.CutPrefix(arg, flagDesktop); found {
			return value
		}
	}

	return ""
}

// Prints where the application has installed itself and what starts it with the session. Returns an error if the installation can not be looked at.
func printStatus() error {
	report, err := desktop.Status()
	if err != nil {
		return err
	}

	fmt.Print(report)

	return nil
}
