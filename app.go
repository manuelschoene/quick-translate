package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.listenForSignals()
	a.start()
}

func (a *App) listenForSignals() {
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Could not start socket:", err.Error())
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

			runtime.WindowShow(a.ctx)

		}
		conn.Close()
	}
}

func (a *App) start() {
	cp, err := ClipboardInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	content := GetTranslatableContent(cp)

	pt := PendingTranslation{
		Content:    content,
		TargetLang: "de",
	}

	translation, sourceLang := translate(pt)

	runtime.WindowShow(a.ctx)

	Notify(translation, sourceLang, strings.ToUpper(pt.TargetLang), cp)

	runtime.WindowHide(a.ctx)
}

func GetTranslatableContent(cp Clipboard) string {
	out, err := cp.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if out == "" {
		fmt.Println("Error: Clipboard is empty")
		os.Exit(1)
	}

	return out
}

type PendingTranslation struct {
	Content, SourceLang, TargetLang string
}

type Provider interface {
	TranslateText(PendingTranslation) (string, string, error)
}

func translate(pt PendingTranslation) (string, string) {
	provider, err := InitDeepLProvider()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	translation, sourceLang, err := provider.TranslateText(pt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return translation, sourceLang
}