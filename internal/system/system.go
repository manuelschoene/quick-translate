package system

import (
	"log/slog"
	"strings"
	"sync"
)

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

// Creates a new structured logger that has been configured to fit the operating system's logging capabilities. The logger is a singleton, so calling this function multiple times returns the same instance. The logger is configured only once. Uses the default logger from the slog package.
func NewLogger() *slog.Logger {
	sync.OnceFunc(configureLogger)()

	return slog.Default()
}
