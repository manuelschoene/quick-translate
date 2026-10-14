package system

import (
	"errors"
	"os"
)

// The variables libc resolves the message locale from, in its order of precedence.
var localeVariables = []string{"LC_ALL", "LC_MESSAGES", "LANG"}

// Reads the locale from the first set variable of localeVariables.
func locale() (string, error) {
	for _, name := range localeVariables {
		if value := os.Getenv(name); len(value) != 0 {
			return normalizeLocale(value)
		}
	}

	return "", errors.New("system: locale not set, none of LC_ALL, LC_MESSAGES or LANG is set")
}
