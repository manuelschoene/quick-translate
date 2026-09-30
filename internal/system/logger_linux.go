package system

import (
	"log"
	"os"
)

// Configures the standard logger to write to stderr without any timestamps or prefixes as systemd does include its own. Does not create the logger instance.
func configureLogger() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}
