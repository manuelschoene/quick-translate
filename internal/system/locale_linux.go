package system

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/language"
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

	return "", errors.New("system: locale not set in LC_ALL, LC_MESSAGES or LANG")
}

// Strips the encoding suffix (e.g. "de_DE.UTF-8") and parses the rest as a BCP 47 tag.
func normalizeLocale(raw string) (string, error) {
	s, _, _ := strings.Cut(raw, ".")

	tag, err := language.Parse(s)
	if err != nil {
		return "", fmt.Errorf("system: parse locale %q: %w", s, err)
	}

	return tag.String(), nil
}
