package system

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/adrg/xdg"
)

// Manages file operations and abstracts the paths of the running operating systems
type FileService struct {
	table []entry
}

// Represents a file or directory with metadata
type entry struct {
	resource Resource // Identifying key
	category category
	name     string // Name displayed to the user, depending on the operating system. Internal entries have no name.
	path     string
	mode     os.FileMode // Targeted permissions. Non-application owned resources have expected values.
	parent   Resource    // Pointer to directory row. Zero when it sits in a directory the application does not own.
}

// File or directory the application reads or writes to. Named after what it holds, as the path changes between operating systems.
type Resource int

const (
	Autostart Resource = iota + 1 // Counted from one, so the zero value marks a non-application owned directory, which is not a resource
	CacheRoot
	ConfigRoot
	DataRoot
	History
	Icon
	Languages
	Launcher
	Settings
	WindowRules
)

// Defines the handling of the file regarding reporting, purging and writing.
type category int

const (
	data       category = iota + 1 // User data inside the user's directories; written, reported, removed
	installed                      // Manually installed files; not written, reported, not removed
	internal                       // Lifecycle managed by the system/session; written, not reported, not removed
	merged                         // Merged into user-owned files; written, reported, not removed
	registered                     // System integration inside user directories; written, reported, removed
)

// Used for path creation and directory names
const appSlug = "quick-translate"

// Returned if the requested resource is not available on the running operating system.
var errNoRow = errors.New("Quick Translate does not know where to keep one of its files on this operating system.")

// Creates a new file service configured with the paths for the running operating system. Resolves the base paths once as they do not change.
func NewFileService() *FileService {
	return &FileService{table: entries()}
}

// Returns the location of a resource with the given parts appended. Returns an empty path for a non-existing resource on the running operating system.
func (f *FileService) Path(r Resource, parts ...string) string {
	row := f.row(r)
	if row == nil {
		return ""
	}

	return filepath.Join(append([]string{row.path}, parts...)...)
}

// Reports whether a resource is in place. A path that can not be read at all counts as missing.
func (f *FileService) Exists(r Resource, parts ...string) bool {
	path := f.Path(r, parts...)
	if len(path) == 0 {
		return false
	}

	_, err := os.Stat(path)

	return err == nil
}

// Reads a resource. A file that does not exist is answered as empty without error. Returns an error on read-failure.
func (f *FileService) Read(r Resource, parts ...string) ([]byte, error) {
	row := f.row(r)
	if row == nil {
		return nil, errNoRow
	}

	path := f.Path(r, parts...)

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, row.failure("read", path, err)
	}

	return content, nil
}

// Writes a resource with targeted permissions while creating missing directories to it. Returns an error on write-failure.
func (f *FileService) Write(r Resource, content []byte, parts ...string) error {
	row := f.row(r)
	if row == nil {
		return errNoRow
	}

	path := f.Path(r, parts...)

	if err := f.ensureDirectoryOf(row, path); err != nil {
		return err
	}

	if err := os.WriteFile(path, content, row.modeFor(path)); err != nil {
		return row.failure("write", path, err)
	}

	return nil
}

// Creates the directories to a resource on demand. If the directory already exists, nothing is done. Does not change permissions of existing directories.
func (f *FileService) EnsureDirectory(r Resource, parts ...string) error {
	row := f.row(r)
	if row == nil {
		return errNoRow
	}

	path := f.Path(r, parts...)

	if err := os.MkdirAll(path, row.mode); err != nil {
		return fmt.Errorf("Could not create the directory '%s': %w", path, err)
	}

	return nil
}

// Creates the file on demand. For an existing file, narrows the permissions of the path and returns true. For a missing file, the content is written and false is returned. Returns an error if the file can not be written.
func (f *FileService) EnsureFile(r Resource, content []byte, parts ...string) (bool, error) {
	if !f.Exists(r, parts...) {
		return true, f.Write(r, content, parts...)
	}

	f.Narrow(r, parts...)

	if row := f.row(r); row != nil && row.parent != 0 {
		f.Narrow(row.parent)
	}

	return false, nil
}

// Narrows the permissions of a resource to the targeted ones. A path that is narrow enough is not modified. Failing to narrow reports the error and continues, as the application can still run.
func (f *FileService) Narrow(r Resource, parts ...string) {
	row := f.row(r)
	if row == nil {
		return
	}

	path := f.Path(r, parts...)

	info, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := row.modeFor(path)

	if info.Mode().Perm()&^mode.Perm() == 0 {
		return
	}

	if err := os.Chmod(path, mode); err != nil {
		fmt.Printf("Could not narrow the permissions of '%s': %v\n", path, err)
		return
	}

	fmt.Printf("Narrowed the permissions of '%s', which is not meant to be readable by other users.\n", path)
}

// Removes a resource and reports whether it existed. Merged and installed categories are not removed. Checks if directories are owned by the application to avoid removing unexpected directories. Returns an error if the resource cannot be removed.
func (f *FileService) Remove(r Resource, parts ...string) (bool, error) {
	row := f.row(r)
	if row == nil {
		return false, errNoRow
	}

	if row.category == installed || row.category == merged {
		return false, fmt.Errorf("Refusing to remove the %s at '%s', which is not Quick Translate's to remove.", row.name, row.path)
	}

	path := f.Path(r, parts...)

	info, err := os.Stat(path)
	if err != nil {
		return false, nil
	}

	if !info.IsDir() {
		if err := os.Remove(path); err != nil {
			return false, row.failure("remove", path, err)
		}

		return true, nil
	}

	if filepath.Base(path) != appSlug {
		return false, fmt.Errorf("Refusing to remove '%s', which is not a directory of Quick Translate.", path)
	}

	if err := os.RemoveAll(path); err != nil {
		return false, row.failure("remove", path, err)
	}

	return true, nil
}

// Returns the row of a resource, or nil when the running operating system has none for it.
func (f *FileService) row(r Resource) *entry {
	for i := range f.table {
		if f.table[i].resource == r {
			return &f.table[i]
		}
	}

	return nil
}

// Creates the parent directory the given path sits in, with its corresponding permissions. Defaults to 0755 when not defined by the row.
func (f *FileService) ensureDirectoryOf(row *entry, path string) error {
	mode := os.FileMode(0755)

	switch {
	case path != row.path:
		mode = row.mode
	case row.parent != 0:
		if parent := f.row(row.parent); parent != nil {
			mode = parent.mode
		}
	}

	directory := filepath.Dir(path)

	if err := os.MkdirAll(directory, mode); err != nil {
		return fmt.Errorf("Could not create the directory '%s': %w", directory, err)
	}

	return nil
}

// Returns the resources of the given categories, in the order of the table.
func (f *FileService) resources(categories ...category) []Resource {
	found := make([]Resource, 0, len(f.table))

	for _, row := range f.table {
		if slices.Contains(categories, row.category) {
			found = append(found, row.resource)
		}
	}

	return found
}

// Returns all files of the given categories in order of the table. Merged categories report whether the file is present, but not whether the application's part is in place.
func (f *FileService) report(categories ...category) []File {
	found := f.resources(categories...)

	files := make([]File, 0, len(found))
	for _, resource := range found {
		files = append(files, f.file(resource, f.Exists(resource)))
	}

	return files
}

// Returns a file description for the given resource with its presence.
func (f *FileService) file(r Resource, present bool) File {
	row := f.row(r)
	if row == nil {
		return File{}
	}

	return File{Name: row.name, Path: row.path, Present: present}
}

// Returns the permissions the thing at the given path is meant to have. A row's own path gets the row's mode; a file inside a directory row gets that directory's permissions without the execute bits.
func (e *entry) modeFor(path string) os.FileMode {
	if path == e.path {
		return e.mode
	}

	return e.mode & 0666
}

// Returns the error of an operation on the row, which names the file when the user knows it by a name and gives the path alone otherwise.
func (e *entry) failure(verb string, path string, err error) error {
	if len(e.name) == 0 {
		return fmt.Errorf("Could not %s '%s': %w", verb, path, err)
	}

	return fmt.Errorf("Could not %s the %s at '%s': %w", verb, e.name, path, err)
}

// The entries every operating system shares, built from base directories. Operating systems that have different base directories must define their own entries() instead.
func userEntries() []entry {
	cacheRoot := filepath.Join(xdg.CacheHome, appSlug)
	configRoot := filepath.Join(xdg.ConfigHome, appSlug)
	dataRoot := filepath.Join(xdg.DataHome, appSlug)

	return []entry{
		{resource: CacheRoot, name: "Language cache", path: cacheRoot, mode: 0700, category: data},
		{resource: ConfigRoot, name: "Configuration", path: configRoot, mode: 0700, category: data},
		{resource: DataRoot, name: "History", path: dataRoot, mode: 0700, category: data},
		{resource: History, path: filepath.Join(dataRoot, "history.db"), mode: 0600, category: internal, parent: DataRoot},
		{resource: Languages, path: filepath.Join(cacheRoot, "translations"), mode: 0700, category: internal, parent: CacheRoot},
		{resource: Settings, path: filepath.Join(configRoot, "config.yml"), mode: 0600, category: internal, parent: ConfigRoot},
	}
}
