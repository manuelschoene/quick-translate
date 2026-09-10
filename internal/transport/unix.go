//go:build unix

package transport

import (
	"fmt"
	"os"
	"path/filepath"
)

// Returns the path of the socket the running instance listens on. The XDG runtime directory is preferred when it is set, because it belongs to the session alone and is cleaned up on logout. It is exclusive to freedesktop-compliant desktop environment. Falls back to a directory scoped to the current user inside the system's temporary directory, so unrelated users on a shared machine can neither collide on the same socket path nor read or replace each other's socket. Returns an error if the fallback directory can not be created.
func socketPath() (string, error) {
	if dir, ok := os.LookupEnv("XDG_RUNTIME_DIR"); ok && len(dir) > 0 {
		return filepath.Join(dir, socketName), nil
	}

	dir := filepath.Join(os.TempDir(), fmt.Sprintf("quick-translate-%d", os.Getuid()))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("Could not create the fallback directory for the instance socket: %w", err)
	}

	return filepath.Join(dir, socketName), nil
}
