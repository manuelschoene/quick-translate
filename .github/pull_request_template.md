<!--
The title is the one hard requirement: Conventional Commits, e.g. "fix: Fixed the asset upload".
release-please reads it off main to write the changelog and to decide that a release is due.

Everything below is prose on purpose. The diff is already visible next to this — what it cannot show is
why the change looks the way it does, and what you did to convince yourself it works. Delete the sections
that do not apply; a one-line fix does not need all four.
-->

<!--
Open with one or two sentences: what this changes and why it needed changing. Lead with the problem, not
with the diff. For a fix, name the cause here rather than the symptom.
-->

## What was wrong

<!--
Only for a fix, and only when the cause is worth its own section — several causes, or one that took
finding. Write what actually broke and why, so the next person recognises it. Delete for a feature.
-->

## What changed

<!--
Grouped by file or by concern, each with the reason it is done that way rather than another way. Someone
reading this in six months should not have to reconstruct the decision from the code.
-->

## How this was verified

<!--
The important one. There is no test suite in this repository, so this is the only signal a reviewer has
that the change works. Say what you ran, not that it works:

    Ran make install, pressed Meta+T on Wayland/KDE, translated a selection, stepped back through the
    history, confirmed the copy button writes to the clipboard.

beats "tested". Reproducing a bug before the fix and again afterwards is worth writing down.

If the change touches the desktop integration, the clipboard or the installation, name the distribution,
desktop environment and session type (Wayland or X11) you tested on — those paths differ a lot between
setups.
-->

## After merging

<!--
Anything that does not happen by itself: a repository setting to flip, a workflow to trigger by hand, a
follow-up worth an issue. Delete if there is nothing.
-->

---

- [ ] The title is a Conventional Commit
- [ ] I can explain every line of this change
- [ ] The files that describe the changed behaviour are updated — see the table at the end of `CLAUDE.md`
- [ ] `frontend/wailsjs/` is regenerated and committed
