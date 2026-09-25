package cli

import (
	"fmt"
	"io"

	"quick-translate/internal/system"
)

var purgeConfirmation = confirmation{
	question: `Your personal data, including your translation history and provider auth keys are part of this and cannot be
recovered. Quit the application before continuing, otherwise files and modifications will be written again.`,
	prompt:   "Are you sure you want to remove them?",
	fallback: false, // Require user to actively say yes, as files cannot be recovered once they are gone.
}

// Removes all application files and modifications. Asks for user confirmation as the action cannot be undone. Returns an error if the files and modifications can not be determined or removed.
func purge(out interaction, _ string, _ []string) error {
	files := system.NewFileService()

	removals, err := system.Removals(files)
	if err != nil {
		return err
	}

	if len(removals) == 0 {
		fmt.Fprintln(out, "Nothing to be removed, no files and modifications have been written yet.")

		return nil
	}

	printRemovals(out, "The following files and modifications will be removed:", removals)

	if !out.confirm(purgeConfirmation) {
		fmt.Fprintln(out, "Cancelled, nothing was removed.")

		return nil
	}

	fmt.Fprintln(out)

	removed, err := system.Purge(files)

	printRemovals(out, "Removed:", removed)
	fmt.Fprintln(out)

	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Successfully removed all files and modifications.")

	return nil
}

// Prints the given files with their name and path in form of a list, below the given heading.
func printRemovals(out io.Writer, heading string, removals []system.File) {
	fmt.Fprintln(out, heading)
	fmt.Fprintln(out)

	for _, file := range removals {
		fmt.Fprintf(out, "  %-15s %s\n", file.Name, file.Path)
	}
}
