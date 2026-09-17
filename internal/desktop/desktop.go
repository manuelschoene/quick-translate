package desktop

import (
	"errors"
	"fmt"
	"strings"
)

// The name the desktop matches this application on, in the reverse domain form GTK and D-Bus require. The
// application hands it to Wails as the GTK application id, the desktop entry and the icon are named after
// it, and KWin matches its rules against it, so every one of those has to read it from here. GTK would
// otherwise derive 'org.wails.quick_translate' from the application name, under a domain this project does
// not own.
const ApplicationID = "io.github.manuelschoene.QuickTranslate"

// The desktop environments an installation can be tailored to. Generic uses only what every freedesktop-compliant desktop understands and is what every environment without a case of its own falls back to, while KDE adds the parts Plasma offers on top of it.
const (
	EnvironmentGeneric = "generic"
	EnvironmentKDE     = "kde"
)

// Returned when the operating system the application runs on has no installation of its own yet.
var ErrUnsupported = errors.New("Quick Translate can not install itself on this operating system yet. Please start the application by hand and bind it to a shortcut in the settings of your desktop.")

// One file the installation owns, reported by the status command so the user can see where the application has put itself.
type File struct {
	Purpose string
	Path    string
	Present bool
}

// What the installation looks like on this machine: the desktop it was detected for, the binary that is answering, how the application is started with the session and every file the installation owns.
type Report struct {
	Environment string
	Executable  string
	Autostart   string
	Files       []File
}

// Renders the report as the block of text the status command prints, with one line per file that tells whether it is installed.
func (r *Report) String() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Desktop:    %s\n", r.Environment)
	fmt.Fprintf(&out, "Executable: %s\n", r.Executable)
	fmt.Fprintf(&out, "Autostart:  %s\n\n", r.Autostart)

	for _, file := range r.Files {
		state := "missing"
		if file.Present {
			state = "installed"
		}

		fmt.Fprintf(&out, "  %-9s  %-13s %s\n", state, file.Purpose, file.Path)
	}

	return out.String()
}

// Registers the running binary with the desktop of the operating system, which means its icon in every size it was rendered in, a desktop entry and whatever starts it with the graphical session. The environment picks the flavour of the installation and is detected from the session when it is empty. Returns an error if a file can not be written or the operating system has no installation of its own yet.
func Install(environment string) error {
	return install(environment)
}

// Removes everything Install has put into the desktop, but not the binary itself, which is owned by whoever has copied it into place. Files that are already gone are not an error, so a partial installation can be cleaned up as well. Returns an error only if a file that is there can not be removed.
func Uninstall() error {
	return uninstall()
}

// Removes the files the application has written for the user: the configuration, the history of everything that was translated and the cached language lists. Deliberately separate from Uninstall, which never touches them, so that reinstalling or rebuilding keeps everything the user has set up and only an explicit request throws it away. Returns an error if the directories can not be determined or the operating system has no installation of its own yet.
func Purge() error {
	return purge()
}

// Reports where the application has installed itself and what starts it with the session. Returns an error if the directories of the desktop can not be determined or the operating system has no installation of its own yet.
func Status() (*Report, error) {
	return status()
}
