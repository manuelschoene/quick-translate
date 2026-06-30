package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	// "github.com/wailsapp/wails/v2/pkg/runtime"
	"quick-translate/internal/clipboard"
	"quick-translate/internal/provider"
)

type App struct {
	ctx context.Context
	provider *provider.ProviderService
	clipboard *clipboard.ClipboardService
	translation *provider.Translation
	tmux sync.Mutex
}

func NewApp(ps *provider.ProviderService, cs *clipboard.ClipboardService) *App {
	t := &provider.Translation{
		SourceLang: &provider.Language{
			Key: "en",
			Name: "English",
			Source: true,
			Target: true,
		},
		TargetLang: &provider.Language{
			Key: "de",
			Name: "German",
			Source: true,
			Target: true,
		},
	}

	return &App{
		provider:  ps,
		clipboard: cs,
		translation: t,
		tmux: sync.Mutex{},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go a.listenOnSocket()
}

func (a *App) listenOnSocket() {
	// Remove existing socket file if app was not shut down properly
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Could not start socket: ", err.Error())
		return
	}
	defer listener.Close()

	// Wait for incoming signal from other instance
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		buf := make([]byte, 4)
		_, err = conn.Read(buf)
		if err == nil && string(buf) == "show" {

			fmt.Println("Running translation...")

		}
		conn.Close()
	}
}