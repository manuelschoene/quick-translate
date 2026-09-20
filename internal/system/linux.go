//go:build linux

package system

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

// Removes all files and modifications the application has written for the user. Returns the files that were present and removed and an error containing all remove failures. The KWin rules are only removed if they were installed and KWin is reconfigured. Does not fail on the first error to try to remove as many files as possible.
func Purge(files *FileService) ([]File, error) {
	removed := make([]File, 0, len(files.table))
	var failures []error

	for _, resource := range files.resources(data, registered) {
		gone, err := files.Remove(resource)
		if err != nil {
			failures = append(failures, err)
			continue
		}

		if gone {
			removed = append(removed, files.file(resource, true))
		}
	}

	switch gone, err := removeRules(files); {
	case err != nil:
		failures = append(failures, err)
	case gone:
		removed = append(removed, files.file(WindowRules, true))
		reconfigureKWin()
	}

	return removed, errors.Join(failures...)
}

// Reports which files and directories the purge command would remove. Missing resources are not included in the list. KWin rules are only reported if the are installed inside the file.
func Removals(files *FileService) ([]File, error) {
	all := append(files.report(data, registered), files.file(WindowRules, rulesInstalled(files)))

	var removals []File = make([]File, 0, len(all))
	for _, file := range all {
		if file.Present {
			removals = append(removals, file)
		}
	}

	return removals, nil
}

// Creates a report of the application's integration into the system. Returns an error if the running binary cannot be determined.
func Status(files *FileService) (*Report, error) {
	executable, err := executablePath()
	if err != nil {
		return nil, err
	}

	return &Report{
		Executable:  executable,
		Environment: os.Getenv("XDG_CURRENT_DESKTOP"),
		Integration: append(files.report(installed, registered), files.file(WindowRules, rulesInstalled(files))),
		UserFiles:   files.report(data),
	}, nil
}

// Returns the path of the binary that is running, with symbolic links resolved so the report names the file itself rather than a link that may be replaced later. Returns an error if the operating system does not answer with a path.
func executablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Could not determine the path of the running binary: %w", err)
	}

	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		return resolved, nil
	}

	return executable, nil
}

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

// Runs a command without treating failures as errors. The command is only run if it is installed and errors are reported without stopping the application.
func runOptional(name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}

	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		fmt.Printf("Could not run '%s': %v. %s\n", name, err, strings.TrimSpace(string(out)))
	}
}

// All file entries for Linux: shared entries, desktop session entries and KWin rules. Not written files are necessary for status reports.
func entries() []entry {
	return append(userEntries(),
		entry{
			resource: Autostart,
			name:     "Autostart",
			path:     filepath.Join(xdg.ConfigHome, "autostart", ApplicationID+".desktop"), // Defined by the XDG Autostart specification
			mode:     0644,                                                                 // Expected mode as Wails manages the file
			category: registered,
		},
		entry{
			resource: Icon,
			name:     "Icon",
			path:     filepath.Join(xdg.DataHome, "icons", "hicolor", "scalable", "apps", ApplicationID+".svg"),
			mode:     0644,
			category: installed,
		},
		entry{
			resource: Launcher,
			name:     "Desktop entry",
			path:     filepath.Join(xdg.DataHome, "applications", ApplicationID+".desktop"),
			mode:     0644,
			category: installed,
		},
		entry{
			resource: WindowRules,
			name:     "Window rules",
			path:     filepath.Join(xdg.ConfigHome, "kwinrulesrc"),
			mode:     0644,
			category: merged, // Merged as KWin uses a single INI style file for all rules
		},
	)
}
