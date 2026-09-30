package system

import (
	"os"
	"strings"
)

// Integrates with the part of the system the user owns. For KDE Plasma, window rules are merged into KWin's settings. For other desktops, there is nothing to integrate yet. Returns an error if the rule file cannot be read or written.
func integrate(files *FileService) error {
	if !isKDE() {
		return nil
	}

	return ensureRules(files)
}

// Reports whether the current desktop session is KDE Plasma. Needed for window rules integration specific to KWin. XDG_CURRENT_DESKTOP is a colon-separated list of environment names, and the desktop sets it to "KDE" for Plasma. The check is case-insensitive.
func isKDE() bool {
	for name := range strings.SplitSeq(os.Getenv("XDG_CURRENT_DESKTOP"), ":") {
		if strings.EqualFold(name, "KDE") {
			return true
		}
	}

	return false
}
