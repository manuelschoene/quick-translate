package main

import (
	"embed"
	"fmt"
	"net"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS
const socketPath = "/tmp/quick-translate.sock"

func main() {
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		// If we can connect to the socket, it means another instance is running, shutting down this instance and sending a message to the running instance
		conn.Write([]byte("show"))
		conn.Close()
		os.Exit(0)
	}

	// If we can't connect to the socket, it means no other instance is running, so we can start this instance
	app := NewApp()

	appErr := wails.Run(&options.App{
		Title:  "Quick Translate",
		Width:  400,
		Height: 250,
		DisableResize: true,
		Frameless: true,
		StartHidden: true,
		AlwaysOnTop: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: true,
		},
	})

	if appErr != nil {
		fmt.Println("Error:", appErr.Error())
	}
}
