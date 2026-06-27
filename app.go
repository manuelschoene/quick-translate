package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	start()
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func start() {
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

	Notify(translation, sourceLang, strings.ToUpper(pt.TargetLang), cp)
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