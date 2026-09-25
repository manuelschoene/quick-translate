package cli

import (
	"fmt"
	"io"

	"quick-translate/internal/system"
)

// Prints the application's integration into the system. Installed files and modifications are listed with their state, name and path.
func printStatus(out interaction, _ string, _ []string) error {
	files := system.NewFileService()

	report, err := system.Status(files)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Executable:      %s\n", report.Executable)
	fmt.Fprintf(out, "Desktop session: %s\n", session(report))

	printFiles(out, "Desktop integration", "installed", report.Integration)
	printFiles(out, "Your personal data", "present", report.UserFiles)

	return nil
}

// Returns the label of the desktop session. If no desktop session is detected, a sensible default is returned.
func session(report *system.Report) string {
	if len(report.Environment) == 0 {
		return "unknown"
	}

	return report.Environment
}

// Prints a section with heading and a list of files, each with its state, name and path. present is the label for a file that is present.
func printFiles(out io.Writer, heading string, present string, files []system.File) {
	fmt.Fprintf(out, "\n%s:\n", heading)

	for _, file := range files {
		state := "missing"
		if file.Present {
			state = present
		}

		fmt.Fprintf(out, "  %-*s  %-15s %s\n", 9, state, file.Name, file.Path)
	}
}
