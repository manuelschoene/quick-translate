package main

import (
	"log"
	"net"
	"os"

	"quick-translate/internal/clipboard"
	"quick-translate/internal/provider"
)


func main() {
	connectToSocket()
	app := wireServices()
	runApp(app)	
}

const socketPath = "/tmp/quick-translate.sock"

func connectToSocket() {
	conn, err := net.Dial("unix", socketPath)

	// If we can connect to the socket, it means another instance is running, shutting down this instance and sending a message to the running instance
	if err == nil {
		conn.Write([]byte("show"))
		conn.Close()
		os.Exit(0)
	}

	// If we can't connect to the socket, it means no other instance is running, so we can start this instance
}

func wireServices() *App {
	cps, err := clipboard.NewService()
	if err != nil {
		log.Fatal(err.Error())
	}

	ps, err := provider.NewService()
	if err != nil {
		log.Fatal(err.Error())
	}
	
	return NewApp(ps, cps)
}
