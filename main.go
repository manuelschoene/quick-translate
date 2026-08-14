package main

import (
	"log"

	"quick-translate/internal/transport"
)

// Starts the application, or hands the shortcut over to the instance that is already running and stops right away.
func main() {
	if transport.Connect() {
		return
	}

	adapter, err := transport.NewAdapter()
	if err != nil {
		log.Fatalf("Could not start Quick Translate: %v", err)
	}

	runApp(adapter)
}
