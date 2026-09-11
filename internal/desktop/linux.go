//go:build linux

package desktop

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed assets/linux
var assets embed.FS

// The directory inside the embedded file system the templates of this operating system live in.
const assetDir = "assets/linux"

// The templates the installation renders. The desktop entry exists twice, because the KDE one carries the global shortcut Plasma reads from it, which every other desktop ignores.
const (
	entryTemplate    = "quick-translate.desktop"
	kdeEntryTemplate = "kde/quick-translate.desktop"
	serviceTemplate  = "quick-translate.service"
	rulesTemplate    = "kde/kwinrules.ini"
)

// The names the installed files carry. The desktop entry is installed under the same name for both flavours, so switching between them replaces the entry instead of leaving a second one behind.
const (
	appName     = "quick-translate"
	entryFile   = "quick-translate.desktop"
	serviceFile = "quick-translate.service"
)

// The name the icon is installed and looked up under. It deliberately carries no dash: an icon name that
// the current theme does not know is looked up again with everything after its last dash cut off, and that
// happens inside the theme before the next one is tried. 'quick-translate' therefore collapses to Breeze's
// unrelated 'quick' icon and the theme the application installs into is never reached.
const iconName = "quicktranslate"

// The sizes the icon was rendered in and is installed in, which are the ones the freedesktop icon theme
// defines for application icons. A desktop asks for whatever size fits what it draws and how it is scaled,
// so shipping the rendered sizes keeps it from stretching a single file and losing sharpness.
var iconSizes = []int{16, 22, 24, 32, 48, 64, 128, 256}

// The characters that have a meaning of their own inside the 'Exec' of a desktop entry and the 'ExecStart'
// of a systemd unit. A path that carries one of them has to be quoted, and a path that does not must not
// be, because both KDE and GNOME derive the binary a window belongs to from the unquoted form and quotes
// that are not needed keep them from recognizing it.
const reservedInExec = " \t\n\"'\\><~|&;$*?#()`"

// The global shortcut the KDE entry is bound to. Plasma is the only desktop that can be told about a shortcut through the desktop entry; every other one leaves this to the user.
const kdeShortcut = "Meta+T"

// The directories of the desktop the installation writes into, resolved once so every step works on the same set of paths.
type paths struct {
	applications string
	iconTheme    string
	systemd      string
	autostart    string
	kwinRules    string

	// The directories the application writes its own files into, which the purge removes. They are named
	// after the application in the three freedesktop base directories, which is how 'config', 'history' and
	// 'language' build their paths as well. A package that ever moves its files elsewhere has to be followed
	// here, because this is the only place that knows about all three of them at once.
	appConfig string
	appData   string
	appCache  string
}

// The values the templates are rendered with: the binary the desktop entry and the service start, the name the icon is looked up under in the theme, the window class KWin matches its rules against and the shortcut the KDE entry is bound to.
type values struct {
	Exec     string
	Icon     string
	Class    string
	Shortcut string
}

// Resolves the directories of the desktop from the freedesktop base directories, which is what XDG_CONFIG_HOME and XDG_DATA_HOME point at when they are set. Returns an error if neither the environment nor the home directory answers.
func newPaths() (*paths, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("Could not determine the configuration directory: %w", err)
	}

	data, err := dataDir()
	if err != nil {
		return nil, err
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("Could not determine the cache directory: %w", err)
	}

	theme := filepath.Join(data, "icons", "hicolor")

	return &paths{
		applications: filepath.Join(data, "applications"),
		iconTheme:    theme,
		systemd:      filepath.Join(config, "systemd", "user"),
		autostart:    filepath.Join(config, "autostart"),
		kwinRules:    filepath.Join(config, "kwinrulesrc"),

		appConfig: filepath.Join(config, appName),
		appData:   filepath.Join(data, appName),
		appCache:  filepath.Join(cache, appName),
	}, nil
}

// Returns the file one size of the icon is installed as, inside the directory the freedesktop layout gives
// that size within the icon theme.
func iconTarget(dirs *paths, size int) string {
	return filepath.Join(dirs.iconTheme, fmt.Sprintf("%dx%d", size, size), "apps", iconName+".png")
}

// Returns the directory user-specific data files are stored in, which is what XDG_DATA_HOME points at when it is set and '~/.local/share' otherwise. The standard library has no counterpart to os.UserConfigDir for this directory, so the freedesktop rule is applied here. Returns an error if the home directory can not be determined.
func dataDir() (string, error) {
	if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok && len(dir) > 0 {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Could not determine the home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share"), nil
}

// Returns the desktop environment the installation is tailored to. A choice that was given on the command line wins, an empty one is detected from XDG_CURRENT_DESKTOP, which the desktop sets to a colon-separated list of the environments it identifies as. Everything that is not Plasma is treated as generic.
func resolveEnvironment(choice string) string {
	if len(choice) > 0 {
		if strings.EqualFold(choice, EnvironmentKDE) {
			return EnvironmentKDE
		}

		return EnvironmentGeneric
	}

	for name := range strings.SplitSeq(os.Getenv("XDG_CURRENT_DESKTOP"), ":") {
		if strings.EqualFold(name, "KDE") {
			return EnvironmentKDE
		}
	}

	return EnvironmentGeneric
}

// Returns the path of the binary that is running, with symbolic links resolved so the desktop entry and the service point at the file itself rather than at a link that may be replaced later. Returns an error if the operating system does not answer with a path.
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

// Returns the path of the binary the way it is written into the desktop entry and the service unit, which
// is the path itself whenever it can be. Only a path that carries a character with a meaning of its own is
// wrapped in quotes, with the characters that keep their meaning inside quotes escaped, so a home directory
// with a space in it still works.
func quoteExecutable(executable string) string {
	if !strings.ContainsAny(executable, reservedInExec) {
		return executable
	}

	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`).Replace(executable)

	return `"` + escaped + `"`
}

// Renders one of the embedded templates with the values of this installation. Returns an error if the template is missing or can not be filled, which both mean the binary itself is broken.
func render(name string, vals *values) ([]byte, error) {
	body, err := assets.ReadFile(path.Join(assetDir, name))
	if err != nil {
		return nil, fmt.Errorf("Could not read the embedded template '%s': %w", name, err)
	}

	parsed, err := template.New(path.Base(name)).Parse(string(body))
	if err != nil {
		return nil, fmt.Errorf("Could not parse the embedded template '%s': %w", name, err)
	}

	var out bytes.Buffer
	if err := parsed.Execute(&out, vals); err != nil {
		return nil, fmt.Errorf("Could not fill the embedded template '%s': %w", name, err)
	}

	return out.Bytes(), nil
}

// Writes one of the installed files and creates the directories leading to it. Returns an error if the directory or the file can not be written, because a missing file means the installation is incomplete.
func writeFile(target string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("Could not create the directory '%s': %w", filepath.Dir(target), err)
	}

	if err := os.WriteFile(target, content, 0644); err != nil {
		return fmt.Errorf("Could not write '%s': %w", target, err)
	}

	return nil
}

// Removes one of the installed files and reports what happened. A file that is already gone is not an error, because the uninstall must also be able to clean up an installation that was never finished.
func removeFile(target string, purpose string) {
	if err := os.Remove(target); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Could not remove '%s': %v\n", target, err)
		}

		return
	}

	fmt.Printf("    Removed %-13s %s\n", purpose, target)
}

// Runs a program that only makes the desktop notice the installation right away, such as a cache that is rebuilt. A program that is not installed is skipped and a failing one is reported without being treated as an error, because the desktop picks the change up at the next login either way.
func runOptional(name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}

	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		fmt.Printf("Could not run '%s': %v. %s\n", name, err, strings.TrimSpace(string(out)))
	}
}

// Installs the application into the desktop: its icon, the desktop entry it is launched through, whatever starts it with the session and, on Plasma, the window rules its frameless window needs. Returns an error as soon as one of these can not be written, so a broken installation is never reported as a working one.
func install(environment string) error {
	environment = resolveEnvironment(environment)

	executable, err := executablePath()
	if err != nil {
		return err
	}

	dirs, err := newPaths()
	if err != nil {
		return err
	}

	vals := &values{Exec: quoteExecutable(executable), Icon: iconName, Class: appName, Shortcut: kdeShortcut}

	fmt.Printf("==> Installing Quick Translate for the '%s' desktop...\n", environment)
	fmt.Printf("    Binary        %s\n", executable)

	if err := installIcon(dirs); err != nil {
		return err
	}

	if err := installEntry(dirs, vals, environment); err != nil {
		return err
	}

	if err := installAutostart(dirs, vals); err != nil {
		return err
	}

	if environment == EnvironmentKDE {
		if err := installRules(dirs, vals); err != nil {
			return err
		}
	}

	printHint(environment, executable)

	return nil
}

// Writes the application icon into the user's icon theme, so the desktop entry can look it up under its name instead of pointing at an absolute path and every size the desktop asks for is served from the theme. Returns an error if the icon can not be written.
func installIcon(dirs *paths) error {
	for _, size := range iconSizes {
		body, err := assets.ReadFile(fmt.Sprintf("%s/icons/%d.png", assetDir, size))
		if err != nil {
			return fmt.Errorf("Could not read the embedded icon for the size %d: %w", size, err)
		}

		if err := writeFile(iconTarget(dirs, size), body); err != nil {
			return err
		}
	}

	fmt.Printf("    Icons         %s (%d sizes)\n", dirs.iconTheme, len(iconSizes))
	refreshIcons(dirs)

	return nil
}

// Writes the desktop entry the application is launched through. KDE gets an entry of its own, because it carries the global shortcut Plasma binds the application to, and Plasma is told to read its entries again so the shortcut works without a logout. Returns an error if the entry can not be rendered or written.
func installEntry(dirs *paths, vals *values, environment string) error {
	name := entryTemplate
	if environment == EnvironmentKDE {
		name = kdeEntryTemplate
	}

	body, err := render(name, vals)
	if err != nil {
		return err
	}

	target := filepath.Join(dirs.applications, entryFile)
	if err := writeFile(target, body); err != nil {
		return err
	}

	fmt.Printf("    Entry         %s\n", target)
	runOptional("update-desktop-database", dirs.applications)

	if environment == EnvironmentKDE {
		refreshPlasma()
	}

	return nil
}

// Makes the application start with the graphical session. A systemd user service is preferred, because it restarts the application when it fails; when no user manager answers, an autostart entry is written instead, which every freedesktop-compliant desktop understands. Returns an error if neither of the two can be installed.
func installAutostart(dirs *paths, vals *values) error {
	if hasSystemd() {
		return installService(dirs, vals)
	}

	return installAutostartEntry(dirs, vals)
}

// Installs the systemd user service and starts it right away, so the application is available for the shortcut without a logout. Whether systemd accepts the unit is not decided here: the manager reads its unit search path when it starts, so a session whose XDG_CONFIG_HOME differs from the one the manager saw can not find a unit written here at all. That costs the autostart and nothing else — the icon and the desktop entry installed before it keep working and the window rules still follow — so a refusal is reported together with the commands to repeat by hand instead of ending the installation. Returns an error only if the unit itself can not be written.
func installService(dirs *paths, vals *values) error {
	body, err := render(serviceTemplate, vals)
	if err != nil {
		return err
	}

	target := filepath.Join(dirs.systemd, serviceFile)
	if err := writeFile(target, body); err != nil {
		return err
	}

	fmt.Printf("    Autostart     %s\n", target)

	if err := enableService(); err != nil {
		fmt.Printf("\n%v\n", err)
		fmt.Printf("The unit is in place, so only the autostart is missing. Set it up with:\n\n")
		fmt.Printf("    systemctl --user daemon-reload\n")
		fmt.Printf("    systemctl --user enable --now %s\n\n", serviceFile)
	}

	return nil
}

// Asks systemd to read its units again and to start the service now and with every session that follows. Returns the error systemd reported, which carries the reason the caller passes on to the user.
func enableService() error {
	if err := runService("daemon-reload"); err != nil {
		return err
	}

	return runService("enable", "--now", serviceFile)
}

// Installs an autostart entry as the fallback for a session without a systemd user manager. The application is not restarted when it fails in this case, which is the price for an autostart that works on every desktop. Returns an error if the entry can not be rendered or written.
func installAutostartEntry(dirs *paths, vals *values) error {
	body, err := render(entryTemplate, vals)
	if err != nil {
		return err
	}

	// GNOME reads this key to decide whether it starts an autostart entry at all and treats a missing one as disabled in some versions.
	body = append(body, []byte("X-GNOME-Autostart-enabled=true\n")...)

	target := filepath.Join(dirs.autostart, entryFile)
	if err := writeFile(target, body); err != nil {
		return err
	}

	fmt.Printf("    Autostart     %s\n", target)

	return nil
}

// Removes everything the installation has put into the desktop. Every step is carried out even when an earlier one has found nothing, because an installation may have been switched between the flavours or interrupted halfway through. Returns an error only if the window rules can not be read or written.
func uninstall() error {
	dirs, err := newPaths()
	if err != nil {
		return err
	}

	fmt.Println("==> Removing Quick Translate from the desktop...")

	removeService(dirs)
	removeFile(filepath.Join(dirs.autostart, entryFile), "autostart")
	removeFile(filepath.Join(dirs.applications, entryFile), "entry")
	removeIcons(dirs)

	runOptional("update-desktop-database", dirs.applications)
	refreshIcons(dirs)

	removed, err := removeRules(dirs)
	if err != nil {
		return err
	}

	if removed {
		fmt.Printf("    Removed %-13s %s\n", "window rules", dirs.kwinRules)
		reconfigureKWin()
	}

	return nil
}

// Removes the icon in every size it was installed in and reports it as one line, because the sizes are an
// implementation detail of the icon theme rather than something the user installed one by one. The size
// directories themselves are left alone, as every other application installs its icons into them as well.
func removeIcons(dirs *paths) {
	removed := false

	for _, size := range iconSizes {
		target := iconTarget(dirs, size)

		if err := os.Remove(target); err == nil {
			removed = true
		} else if !os.IsNotExist(err) {
			fmt.Printf("Could not remove '%s': %v\n", target, err)
		}
	}

	if removed {
		fmt.Printf("    Removed %-13s %s\n", "icons", dirs.iconTheme)
	}
}

// Stops and disables the systemd user service before its unit is removed. A session without a systemd user manager only has an autostart entry, which is removed by the caller, so nothing is done here.
func removeService(dirs *paths) {
	if !hasSystemd() {
		return
	}

	if err := runService("disable", "--now", serviceFile); err != nil {
		fmt.Println(err)
	}

	removeFile(filepath.Join(dirs.systemd, serviceFile), "service")

	if err := runService("daemon-reload"); err != nil {
		fmt.Println(err)
	}
}

// Removes the directories the application has written its own files into: the configuration, the history of everything that was translated and the cached language lists. Never part of an uninstall on its own, because a reinstall is far more common than a farewell and nobody expects their history to go with the binary. Returns an error only if the directories can not be determined; a directory that can not be removed is reported and the rest are still tried.
func purge() error {
	dirs, err := newPaths()
	if err != nil {
		return err
	}

	fmt.Println("==> Removing the files Quick Translate has written...")

	for _, dir := range []string{dirs.appConfig, dirs.appData, dirs.appCache} {
		removeDirectory(dir)
	}

	return nil
}

// Removes one of the directories the application owns, with everything below it. A directory that is not there is the normal state of an installation that never wrote anything, so it is passed over in silence. The name is checked before the removal, because a base directory that resolved to something unexpected must not take a whole home directory with it.
func removeDirectory(dir string) {
	if filepath.Base(dir) != appName {
		fmt.Printf("Refusing to remove '%s', which is not a directory of Quick Translate.\n", dir)
		return
	}

	if _, err := os.Stat(dir); err != nil {
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		fmt.Printf("Could not remove '%s': %v\n", dir, err)
		return
	}

	fmt.Printf("    Removed       %s\n", dir)
}

// Reports where the application has installed itself, which files are in place and what starts it with the session. Returns an error if the directories of the desktop or the running binary can not be determined.
func status() (*Report, error) {
	dirs, err := newPaths()
	if err != nil {
		return nil, err
	}

	executable, err := executablePath()
	if err != nil {
		return nil, err
	}

	return &Report{
		Environment: resolveEnvironment(""),
		Executable:  executable,
		Autostart:   autostartState(dirs),
		Files: []File{
			reportFile("entry", filepath.Join(dirs.applications, entryFile)),
			{Purpose: "icons", Path: dirs.iconTheme, Present: iconsInstalled(dirs)},
			reportFile("service", filepath.Join(dirs.systemd, serviceFile)),
			reportFile("autostart", filepath.Join(dirs.autostart, entryFile)),
			{Purpose: "window rules", Path: dirs.kwinRules, Present: rulesInstalled(dirs)},
		},
	}, nil
}

// Describes one of the installed files for the status report. A path that can not be read at all counts as missing, because that is what it means for the desktop as well.
func reportFile(purpose string, target string) File {
	_, err := os.Stat(target)

	return File{Purpose: purpose, Path: target, Present: err == nil}
}

// Reports whether the icon is installed in every size, so a status that calls the icons installed means the
// desktop finds a rendered one for whatever size it asks for.
func iconsInstalled(dirs *paths) bool {
	for _, size := range iconSizes {
		if _, err := os.Stat(iconTarget(dirs, size)); err != nil {
			return false
		}
	}

	return true
}

// Describes what starts the application with the session, which is the state of the systemd user service when there is a user manager and the autostart entry otherwise. Meant for the status report, so a state that can not be read is described rather than reported as an error.
func autostartState(dirs *paths) string {
	if !hasSystemd() {
		return "autostart entry, because no systemd user manager answers"
	}

	enabled := serviceProperty("is-enabled")
	if enabled == "not-found" || len(enabled) == 0 {
		// The unit being on disk while systemd does not know it means the manager searches somewhere else,
		// which is worth saying apart from the unit simply never having been installed.
		if _, err := os.Stat(filepath.Join(dirs.systemd, serviceFile)); err == nil {
			return "the systemd unit is installed but the user manager does not see it"
		}

		return "systemd user service is not installed"
	}

	return fmt.Sprintf("systemd user service (%s, %s)", enabled, serviceProperty("is-active"))
}

// Asks systemd for one of the states of the service and answers with what it prints, which it does for a service that is neither enabled nor running as well. An empty answer means systemd could not be asked at all.
func serviceProperty(query string) string {
	out, _ := exec.Command("systemctl", "--user", query, serviceFile).Output()

	return strings.TrimSpace(string(out))
}

// Reports whether the user's own systemd manager can be reached, which is the case in a regular desktop session on a systemd distribution. Both a missing systemctl and a manager that does not answer count as no systemd, so the installation can fall back to an autostart entry.
func hasSystemd() bool {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}

	return exec.Command("systemctl", "--user", "show-environment").Run() == nil
}

// Runs systemctl for the user's own service manager. Returns an error that carries what systemd has printed, because its message is what tells the user why a service was refused.
func runService(args ...string) error {
	out, err := exec.Command("systemctl", append([]string{"--user"}, args...)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Could not run 'systemctl --user %s': %w. %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}

	return nil
}

// Rebuilds the cache of the icon theme, so the desktop finds the icon right away instead of at the next login. The tool belongs to GTK and is not installed on every desktop, which is why a missing one is skipped.
func refreshIcons(dirs *paths) {
	runOptional("gtk-update-icon-cache", "-q", "-t", "-f", dirs.iconTheme)
}

// Prints what the user has to do themselves after the installation. Only Plasma can be told about a global shortcut through the desktop entry, so every other desktop is told which command to bind by hand.
func printHint(environment string, executable string) {
	if environment == EnvironmentKDE {
		fmt.Printf("\nQuick Translate is bound to %s. Change the shortcut under System Settings > Keyboard > Shortcuts.\n", kdeShortcut)
		return
	}

	fmt.Printf("\nThere is no way to bind a global shortcut that works on every desktop, so this is left to you.\n")
	fmt.Printf("Bind the following command to a shortcut in the keyboard settings of your desktop:\n\n    %s\n", executable)
}
