# Security Policy

Quick Translate is in **alpha** and maintained by one person in their spare time. Reports are taken
seriously all the same.

## Supported versions

Only the latest release is supported. Fixes land on `main` and go out with the next release. Run
`quick-translate --version` to see which build you have.

## Reporting a vulnerability

**Email: [schoene-manuel@gmx.de](mailto:schoene-manuel@gmx.de)**

- **Exploitable before a fix exists?** Email first. Please do not open a public issue or pull request.
- **Hardening, a too-wide permission, a dependency bump?** Just
  **[open a pull request](https://github.com/manuelschoene/quick-translate/pulls)**.

Please include what an attacker could achieve, the steps to reproduce (with distribution, desktop and session
type), the output of `quick-translate --version`, and a fix if you have one.

There is no guaranteed response time and no bug bounty. I will acknowledge your report, tell you honestly
whether I consider it a vulnerability, and credit you in the release notes unless you prefer otherwise.

## What Quick Translate does with your data

- **Network:** the selected text goes over HTTPS to the translation provider you configured, and so does the
  request for its language list. Nothing else is sent anywhere.
- **Auth key:** stored **in plain text** in `~/.config/quick-translate/config.yml`. Other users cannot read
  it, but anything running as your user can. Moving it into the OS keyring is planned.
- **History:** every text you translated and its translation, stored **unencrypted** in
  `~/.local/share/quick-translate/history.db`. `history.max_entries: 0` turns it off, and
  `quick-translate --purge` deletes it.
- **Cache:** `~/.cache/quick-translate` only holds the languages a provider supports.
- **Permissions:** every directory above is created with mode `0700`, every file with `0600`.
- **Clipboard:** read and written through an external program (`wl-paste`/`wl-copy`, `xclip` or `xsel`),
  looked up on `PATH`.
- **Session bus:** nothing of yours goes over it. The application owns two names and, under Wayland, holds
  one portal session:
  - `io.github.manuelschoene.QuickTranslate.SingleInstance`, so that a second start reaches the running one
  - `org.kde.StatusNotifierItem-<pid>-1` for the tray icon
  - a session with the desktop portal for the global shortcut

## Scope

**In scope:** anything that lets someone else read your auth key, clipboard or history; lets another local
user drive the running instance over the session bus; turns untrusted input (a translation, a provider
response, a configuration file) into code execution; or sends your data somewhere it should not go.

**Out of scope:**

- The auth key being readable by processes running as your own user (see above)
- Vulnerabilities in the provider's own service
- Anything that requires root on the machine
