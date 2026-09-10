//go:build unix

package language

import (
	"fmt"
	"os"
)

// The environment variables the C library resolves the locale from, in the order it checks them: LC_ALL overrides every category, LC_MESSAGES is the category for the language of message text, and LANG is the general fallback used when neither is set.
var localeEnvVars = []string{"LC_ALL", "LC_MESSAGES", "LANG"}

// Detects the system locale by reading the environment variables the C library resolves it from, in the same precedence order libc uses on every Unix-like operating system (Linux, macOS, the BSDs). This is independent of the distribution or OS, because the precedence is defined by POSIX, not by any single implementation's packaging. Returns an error if none of the variables is set or the value found is not a valid BCP 47 tag.
func locale() (string, error) {
	for _, name := range localeEnvVars {
		value, ok := os.LookupEnv(name)
		if !ok || len(value) == 0 {
			continue
		}

		return normalizeLocale(value)
	}

	return "", fmt.Errorf("Cannot determine system locale: none of LC_ALL, LC_MESSAGES or LANG is set")
}
