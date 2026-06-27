package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
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
