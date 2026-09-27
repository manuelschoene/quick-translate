package system

import "strings"

// The application ID is used by the system to identify the application and match it with the autostart entry, window rules, global shorcut and desktop entry. Reverse DNS syntax is required by GTK and D-Bus.
const ApplicationID = "io.github.manuelschoene.QuickTranslate"

// Name of the application shown to the user and used for system integration.
const AppName = "Quick Translate"

// Used for path creation and directory names
var appSlug = strings.ToLower(strings.ReplaceAll(AppName, " ", "-"))

// The keys the application asks the desktop for.
const Shortcut = "Super+Shift+T"

// Tooltip shown by the desktop for the tray icon. The keys are spelled the way Wails and GNOME name them; KDE Plasma calls the same key Meta.
const trayTooltip = AppName + " — " + Shortcut

type File struct {
	Name    string // What the user knows the file as, e.g., "Desktop entry" or "Configuration"
	Path    string
	Present bool
}

type Report struct {
	Executable  string // Binary path
	Environment string // System or desktop environment name, e.g., "KDE"
	Integration []File // Files that are part of the integration, e.g., desktop entry and window rules
	UserFiles   []File // Files with application data inside the user's home directory, e.g., configuration and history
}
