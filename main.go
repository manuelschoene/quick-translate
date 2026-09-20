package main

import (
	"os"

	"quick-translate/internal/cli"
)

// The version this binary was built as. Stamped in by the build from the version in 'build/config.yml'. Development builds use this default.
var version = "0.0.0-dev"

// Starts the application. A command given on the command line is carried out instead, before anything of the application is set up, so a command is never handed to the instance that is already running.
func main() {
	if handled, code := cli.Run(os.Args[1:], version); handled {
		os.Exit(code)
	}

	runApp()
}
