package transport

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// The name of the socket file the running instance listens on.
const socketName = "quick-translate.sock"

// The message an instance that was started by the shortcut sends to the running one.
const showMessage = "show"

// How long the two instances wait for each other before they give up.
const messageTimeout = 2 * time.Second

// Checks if the application is already running and asks it to show itself. Returns true if another instance has taken the request over, which means this instance is not needed and should stop before it initializes anything.
func Connect() bool {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return false
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(messageTimeout)); err != nil {
		fmt.Printf("Could not set a deadline for the running instance: %v\n", err)
	}

	// The running instance answers the socket, so this instance is not needed either way. Starting a second one would take the socket away from it.
	if _, err := conn.Write([]byte(showMessage)); err != nil {
		fmt.Printf("Could not ask the running instance to show itself: %v\n", err)
	}

	return true
}

// Returns the path of the socket the running instance listens on. The runtime directory of the user is preferred, because it belongs to the session alone and is cleaned up on logout. Falls back to the temporary directory for systems that do not set it.
func socketPath() string {
	dir, ok := os.LookupEnv("XDG_RUNTIME_DIR")
	if !ok || len(dir) == 0 {
		dir = os.TempDir()
	}

	return filepath.Join(dir, socketName)
}

// Starts listening for the instances that are started by the shortcut. The socket file of an instance that did not shut down properly is removed first, which is safe because connecting to it has just failed. Returns an error if the socket can not be created.
func (a *Adapter) listenOnSocket() error {
	path := socketPath()

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not remove the socket of a previous instance: %w", err)
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return fmt.Errorf("Could not listen for the shortcut on '%s': %w", path, err)
	}

	a.listener = listener

	go a.listen()

	return nil
}

// Accepts the instances that are started by the shortcut, one after another, so two presses in a row do not translate at the same time. Runs until the listener is closed by the shutdown.
func (a *Adapter) listen() {
	for {
		conn, err := a.listener.Accept()
		if err != nil {
			// A closed listener is the regular way out of this loop.
			return
		}

		a.handle(conn)
	}
}

// Reads the message of an instance that was started by the shortcut and shows the application if it asks for it.
func (a *Adapter) handle(conn net.Conn) {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(messageTimeout)); err != nil {
		fmt.Printf("Could not set a deadline for the second instance: %v\n", err)
		return
	}

	message := make([]byte, len(showMessage))
	if _, err := io.ReadFull(conn, message); err != nil {
		fmt.Printf("Could not read the message of the second instance: %v\n", err)
		return
	}

	if string(message) != showMessage {
		fmt.Printf("Ignoring the unknown message '%s' of the second instance.\n", message)
		return
	}

	a.show()
}
