package system

// The application ID is used by the system to identify the application and match it with the autostart entry, window rules, global shorcut and desktop entry. Reverse DNS syntax is required by GTK and D-Bus.
const ApplicationID = "io.github.manuelschoene.QuickTranslate"

// The keys the application asks the desktop for. Wails spells the meta key 'Super'.
const Shortcut = "Super+Shift+T"

type File struct {
	Name    string // What the user knows the file as, e.g., "desktop entry" or "configuration"
	Path    string
	Present bool
}

type Report struct {
	Executable  string // Binary path
	Environment string // System or desktop environment name, e.g., "KDE"
	Integration []File // Files that are part of the integration, e.g., desktop entry and window rules
	UserFiles   []File // Files with user data inside the user's home directory, e.g., configuration and history
}
