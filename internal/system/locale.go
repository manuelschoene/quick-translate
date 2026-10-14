package system

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

// Returns the user's locale as a BCP 47 tag, e.g. "de-DE". Fails if the operating system does not reveal one.
//
//wails:ignore
func (i *IntegrationService) Locale() (string, error) {
	return locale()
}

// Strips the encoding suffix (e.g. "de_DE.UTF-8") and parses the rest as a BCP 47 tag.
func normalizeLocale(raw string) (string, error) {
	s, _, _ := strings.Cut(raw, ".")

	tag, err := language.Parse(s)
	if err != nil {
		return "", fmt.Errorf("system: locale %q is not a valid BCP 47 tag", s)
	}

	return tag.String(), nil
}
