//go:build linux

package system

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// The group the window rules are written to. KWin addresses every rule by a name of its own, so a fixed one lets the installation replace and remove exactly its own rule without ever touching the rules the user has made.
const rulesGroup = "8f2c1d64-3a5b-4c7e-9d10-2b6f4a8c0e31"

// The KWin rules with a placeholder at the window class for the application ID
const rules = `Description=Quick Translate
wmclass=%s
wmclasscomplete=false
wmclassmatch=1
above=true
aboverule=3
noborder=true
noborderrule=3
skiptaskbar=true
skiptaskbarrule=3
skippager=true
skippagerrule=3
skipswitcher=true
skipswitcherrule=3
`

// The name the rule carries in the KWin settings. It is what recognizes a rule an earlier version has written under a group name of its own, so such a rule is replaced instead of being left behind as a duplicate.
const rulesDescription = "Quick Translate"

// The group KWin keeps the list of its rules in, and the key inside it that names them in order.
const (
	generalGroup = "General"
	rulesKey     = "rules"
	countKey     = "count"
)

// One group of a KDE settings file: the name it carries in brackets and the lines below it, up to the next group (INI-style).
type group struct {
	name  string
	lines []string
}

// Returns the value of a key inside the group, or an empty string when the group does not carry the key.
func (g *group) value(key string) string {
	for _, line := range g.lines {
		name, value, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(name) == key {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// Sets the value of a key inside the group and appends the key when the group does not carry it yet, so a group written by KWin keeps the order of its own keys.
func (g *group) set(key string, value string) {
	for i, line := range g.lines {
		name, _, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(name) == key {
			g.lines[i] = key + "=" + value
			return
		}
	}

	g.lines = append(g.lines, key+"="+value)
}

// Splits a KDE settings file into its groups. Everything before the first group header is kept as a group without a name, so a file can be written back with nothing but the edited part changed.
func parseGroups(content []byte) []*group {
	groups := []*group{{}}

	for line := range strings.SplitSeq(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			groups = append(groups, &group{name: strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")})
			continue
		}

		current := groups[len(groups)-1]
		current.lines = append(current.lines, line)
	}

	return groups
}

// Renders the groups back into the text of a KDE settings file. Empty lines at the end of a group are dropped and a single one is put between the groups instead, so the file keeps its shape no matter how often it is written.
func renderGroups(groups []*group) []byte {
	var out strings.Builder

	for _, current := range groups {
		lines := trimEmptyLines(current.lines)
		if len(current.name) == 0 && len(lines) == 0 {
			continue
		}

		if out.Len() > 0 {
			out.WriteString("\n")
		}

		if len(current.name) > 0 {
			fmt.Fprintf(&out, "[%s]\n", current.name)
		}

		for _, line := range lines {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}

	return []byte(out.String())
}

// Returns the lines without the empty ones at their end, which is what keeps the rendered file from growing a blank line with every write.
func trimEmptyLines(lines []string) []string {
	end := len(lines)
	for end > 0 && len(strings.TrimSpace(lines[end-1])) == 0 {
		end--
	}

	return lines[:end]
}

// Returns the group with the given name, or nil when the file does not carry it.
func findGroup(groups []*group, name string) *group {
	for _, current := range groups {
		if current.name == name {
			return current
		}
	}

	return nil
}

// Returns the groups without the rule of an earlier installation, both the one under the fixed group name and one an older version has written under a name of its own, which is recognized by its description.
func dropRules(groups []*group) []*group {
	kept := make([]*group, 0, len(groups))

	for _, current := range groups {
		if current.name == rulesGroup || (len(current.name) > 0 && current.value("Description") == rulesDescription) {
			continue
		}

		kept = append(kept, current)
	}

	return kept
}

// Rewrites the list KWin reads its rules from, which lives in the general group and names every rule group in order. The application's own rule is added at the end when it should be installed and left out when it should not, and a name whose group is gone is dropped, because such a rule does not exist for KWin either. The general group is created when the file does not have one yet.
func updateRuleList(groups []*group, include bool) []*group {
	general := findGroup(groups, generalGroup)
	if general == nil {
		general = &group{name: generalGroup}
		groups = append(groups, general)
	}

	names := make([]string, 0, len(groups))
	for name := range strings.SplitSeq(general.value(rulesKey), ",") {
		if len(name) > 0 && name != rulesGroup && findGroup(groups, name) != nil {
			names = append(names, name)
		}
	}

	if include {
		names = append(names, rulesGroup)
	}

	general.set(countKey, strconv.Itoa(len(names)))
	general.set(rulesKey, strings.Join(names, ","))

	return groups
}

// Returns the lines of the rule's group, filled in with the application id KWin matches the window against.
func ruleBody() []string {
	return strings.Split(strings.TrimSpace(fmt.Sprintf(rules, ApplicationID)), "\n")
}

// Ensures the window rules are part of the user's rule file. Earlier versions of the window rules are replaced. KWin is asked to be reconfigured for the rules to take effect without a logout. It is only written when the file does not have the rules installed. Returns an error if the rule file can not be read or written.
func ensureRules(files *FileService) error {
	content, err := files.Read(WindowRules)
	if err != nil {
		return err
	}

	groups := append(dropRules(parseGroups(content)), &group{name: rulesGroup, lines: ruleBody()})

	wanted := renderGroups(updateRuleList(groups, true))
	if bytes.Equal(content, wanted) {
		return nil
	}

	if err := files.Write(WindowRules, wanted); err != nil {
		return err
	}

	fmt.Printf("Merged the window rules of Quick Translate into '%s'.\n", files.Path(WindowRules))
	reconfigureKWin()

	return nil
}

// Removes the window rules from the user's rule file and reports whether the file had to be changed, so the caller only tells KWin about it when something has moved. A rule file that does not exist means there is nothing to remove. Returns an error if the file can not be read or written.
func removeRules(files *FileService) (bool, error) {
	content, err := files.Read(WindowRules)
	if err != nil || len(content) == 0 {
		return false, err
	}

	groups := parseGroups(content)

	kept := dropRules(groups)
	if len(kept) == len(groups) {
		return false, nil
	}

	if err := files.Write(WindowRules, renderGroups(updateRuleList(kept, false))); err != nil {
		return false, err
	}

	return true, nil
}

// Reports whether the application's window rules are part of the user's rule file.
func rulesInstalled(files *FileService) bool {
	content, err := files.Read(WindowRules)
	if err != nil || len(content) == 0 {
		return false
	}

	return findGroup(parseGroups(content), rulesGroup) != nil
}

// Asks KWin to read its settings again, so the window rules take effect without a logout. KWin is only reachable over the session bus of a running Plasma session, which is why a failure is reported and swallowed.
func reconfigureKWin() {
	runOptional("dbus-send", "--session", "--dest=org.kde.KWin", "--type=method_call", "/KWin", "org.kde.KWin.reconfigure")
}
