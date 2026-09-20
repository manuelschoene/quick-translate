package main

import (
	"os"

	"quick-translate/internal/cli"
	"quick-translate/internal/system"
	"quick-translate/internal/transport"
)

// The version this binary was built as. Stamped in by the build from the version in 'build/config.yml'. Development builds use this default.
var version = "0.0.0-dev"

// Starts the application, or hands the shortcut over to the instance that is already running and stops right away. A command given on the command line is carried out instead of any of that, before anything of the application is set up, so a command never reaches the instance that is running.
func main() {
	if handled, code := cli.Run(os.Args[1:], version); handled {
		os.Exit(code)
	}

	files := system.NewFileService()

	if transport.Connect(files) {
		return
	}

	runApp(files)
}
