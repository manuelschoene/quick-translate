package main

import (
	"fmt"
	"os"
)

func main() {
	content := GetTranslatableContent()

	pt := PendingTranslation{
		Content:    content,
		TargetLang: "de",
	}

	fmt.Println("Translating: " + content)

	translation := translate(pt)

	fmt.Println("Translation: " + translation)
}

func GetTranslatableContent() string {
	cp, err := ClipboardInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
	TranslateText(PendingTranslation) (string, error)
}

func translate(pt PendingTranslation) string {
	provider, err := InitDeepLProvider()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	translation, err := provider.TranslateText(pt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return translation
}