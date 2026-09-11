# Security Policy

Quick Translate is in **alpha** and maintained by one person in their spare time. Reports are taken
seriously all the same — please read how to send them below.

## Supported versions

Only the latest version is supported. There are no maintenance branches, and fixes land on `main` and go
out with the next release.

Run `quick-translate --version` to see which build you have.

## Reporting a vulnerability

**Email me: [schoene-manuel@gmx.de](mailto:schoene-manuel@gmx.de)**

If you already have a fix and the problem is not something that could be exploited against people running
the current release — a hardening improvement, a permission that is wider than it needs to be, a dependency
bump — you are equally welcome to just
**[open a pull request](https://github.com/manuelschoene/quick-translate/pulls)** and describe it there.
That is often the fastest route.

For anything an attacker could act on before a fix exists, please email first. A pull request is public the
moment you open it, and the details would be readable by everyone while users are still exposed.

Either way, please do not open a public issue for an exploitable problem.

### What to include

- What the problem is and what an attacker could achieve with it
- The steps to reproduce it, with your distribution, desktop environment and session type
- The output of `quick-translate --version`
- A patch or a suggested fix, if you have one

### What to expect

This is a hobby project, so there is no guaranteed response time and no bug bounty. What I will do:

- Acknowledge your report as soon as I have read it
- Tell you honestly whether I consider it a vulnerability, and why
- Credit you in the release notes when the fix goes out, unless you would rather I did not

## What Quick Translate does with your data

Worth knowing before you report, and worth knowing as a user:

- **Your text goes to the translation provider you configured.** Translating sends the selected text over
  HTTPS to that provider's API — DeepL at the time of writing. Nothing is sent anywhere else, and the
  application makes no other network requests.
- **Your provider auth key is stored in plain text** in `~/.config/quick-translate/config.yml`. The file is
  kept at mode `0600` inside a `0700` directory, so other users on the machine cannot read it, but it is
  not encrypted and anything running as your user can read it. Moving keys into the OS keyring is planned.
- **Your translation history is stored unencrypted** in `~/.local/share/quick-translate/history.db`, at
  mode `0600` in a `0700` directory. It holds every text you translated and its translation. Set
  `history.max_entries` to `0` in the configuration to turn the history off, or remove everything with
  `quick-translate --uninstall --purge`.
- **The clipboard is read through an external program** (`wl-paste`, `xclip` or `xsel`), chosen from your
  session type and looked up on `PATH`.
- **A unix socket** at `$XDG_RUNTIME_DIR/quick-translate.sock` carries the shortcut to the running
  instance. Where `XDG_RUNTIME_DIR` is not set, it falls back to a directory under the system temporary
  directory that is scoped to your user id and created with mode `0700`.

## In scope

Anything that lets someone read your auth keys, your clipboard or your history without already being you on
that machine; anything that lets a second local user talk to the running instance; anything that turns
untrusted input — a translated text, a provider response, a configuration file — into code execution;
and anything that sends your data somewhere it should not go.

## Out of scope

- The auth key being readable by processes running as your own user. That is what the keyring rework is
  for, and it is documented above rather than being a finding.
- Vulnerabilities in the translation provider's own service.
- Anything that requires an attacker to already have root on the machine.
