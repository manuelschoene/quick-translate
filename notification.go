package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func Notify(message string, sourceLang string, targetLang string, cb Clipboard) {
	cmd := exec.Command("notify-send", "Translating from "+sourceLang+" to "+targetLang, message, "--app-name=Quick Translate", "--action=copy=Copy to clipboard", "--expire-time=10000")

	out, err := cmd.Output()
	if err != nil {
		println("Error sending notification:", err.Error())
		return
	}

	action := strings.TrimSpace(string(out))

	fmt.Println(action)

	if action == "copy" {
		err := cb.Write(message)
		if err != nil {
			println("Error copying to clipboard:", err.Error())
		}
	}
}