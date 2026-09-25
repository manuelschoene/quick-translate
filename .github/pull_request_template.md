<!--
Title: a Conventional Commit, e.g. "fix: Fixed the asset upload". release-please builds the changelog from it.
Delete the sections that do not apply; a one-line fix does not need all of them.
-->

<!-- One or two sentences: what this changes and why. For a fix, name the cause rather than the symptom. -->

## What changed

<!-- Grouped by file or concern, each with the reason it is done this way. -->

## How this was verified

<!--
There is no test suite, so this is the reviewer's only signal. Say what you ran, e.g.:

    Built with task archive, installed into ~/.local, pressed Super+Shift+T on
    Wayland/KDE, translated a selection, stepped back through the history.

For the desktop integration, the clipboard or the installation, name distribution, desktop and session type.
-->

## After merging

<!-- Anything that does not happen by itself: a setting, a manual workflow run, a follow-up issue. -->

---

- [ ] The title is a Conventional Commit
- [ ] I can explain every line of this change
- [ ] The docs describing the changed behavior are updated (see the table at the end of `CLAUDE.md`)
- [ ] `frontend/bindings/` is regenerated and committed, if the frontend API changed
