<p align="center">
	<img alt="Quick Translate Logo" src="art/wordmark.svg" width="512">
</p>

<br>

<p align="center">
	<a href="https://www.apache.org/licenses/LICENSE-2.0">
		<img alt="License" src="https://img.shields.io/badge/License-Apache--2.0-blue.svg">
	</a>
	<a href="https://go.dev">
		<img alt="Go" src="https://img.shields.io/badge/Go-1.26.8-00ADD8?logo=go&logoColor=white">
	</a>
	<a href="https://vuejs.org">
		<img alt="Vue.js" src="https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js">
	</a>
</p>

# Quick Translate

> [!WARNING]
> **Quick Translate is in alpha.** Everything described below works, but the application has seen little use outside its own development. Expect rough edges, and expect the configuration format to change without a migration path.
>
> - **Supported:** Linux, on any freedesktop-compliant desktop, with additional integration for KDE Plasma.
> - **Not yet supported:** macOS and Windows.
> - **Installation:** a prebuilt binary from the releases page, or from source. Every release ships two Linux binaries, one for each WebKitGTK version, because a binary linked against one does not start on a system that ships the other.
>
> Problems are very welcome as [GitHub issues](https://github.com/manuelschoene/quick-translate/issues/new/choose) — the bug report form asks for the handful of details that make one reproducible, including the output of `quick-translate --status`.

Quick Translate is a lightweight desktop utility that translates clipboard or selected text on demand, triggered by a single global keyboard shortcut. Translation is performed by a configurable provider, described below.

On Linux it installs itself into any freedesktop-compliant desktop: an icon, a desktop entry and a systemd user service that keeps it running in the background. KDE Plasma additionally gets the global shortcut and the window rules for the popup set up for it.

## Installation

There are two ways in: download a release archive, or build from source. Both end the same way — the binary registers itself with the desktop, which is why the desktop integration behaves identically no matter how the binary got onto the machine.

Run `make help` for the full list of targets and variables.

`main` is the development state and changes as work lands, so what you build from it is effectively a nightly. To build a released version instead, take its source archive from the [releases page](https://github.com/manuelschoene/quick-translate/releases) or check out its tag (`git checkout v0.1.0-alpha`). Either way the build and installation below are the same.

### From a release archive

Each release on the [releases page](https://github.com/manuelschoene/quick-translate/releases) carries two `linux-amd64` archives. Pick the one that matches the WebKitGTK your distribution ships:

```sh
pkg-config --exists webkit2gtk-4.1 && echo webkit41 || echo webkit40
```

`webkit41` fits current distributions (Arch, Fedora 40+, Ubuntu 24.04, Debian 13); `webkit40` fits the older ones that still ship WebKit2GTK 4.0. Both are built on Ubuntu 22.04, so they run on any glibc from 2.35 onwards. Unpack, install the binary, and let it register itself:

```sh
tar -xzf quick-translate-*-linux-amd64-webkit41.tar.gz
cd quick-translate-*-linux-amd64-webkit41
install -D -m 0755 quick-translate ~/.local/bin/quick-translate
~/.local/bin/quick-translate --install
```

Each release also carries a `SHA256SUMS` file, so the download can be checked with `sha256sum --check SHA256SUMS`. You still need a clipboard tool at runtime — see [clipboard access](#clipboard-access) below.

Building from source, described in the rest of this section, is the other route and the one to take if your distribution ships neither WebKitGTK version the releases cover.

### Prerequisites

- `Go` >= 1.26.8, the floor declared in `go.mod`. Earlier 1.26 patches carry standard library vulnerabilities that this project reaches through its HTTPS calls, and the build refuses them. A newer toolchain is downloaded automatically unless you have set `GOTOOLCHAIN=local`
- `Bun` >= 1.3.14, used to install and build the frontend
- `gcc` >= 16.1.1, required to build the CGO-based WebKit bindings used by Wails
- `pkgconf` >= 2.5.1, used by Wails to locate the GTK and WebKit libraries
- `webkit2gtk`, used by Wails to render the frontend on Linux. Both `webkit2gtk-4.1` >= 2.52.5 and the older `webkit2gtk-4.0` work; the `Makefile` asks `pkg-config` which one is installed and sets the `webkit2_41` build tag for the newer one
- `gtk3` >= 3.24.52, used by Wails on Linux
- `upx` >= 5.2.0, used to compress the built binary. Optional: a missing UPX is skipped, and `make build UPX=0` skips it on purpose, which is noticeably faster while developing

The Wails CLI itself is not listed above, as it is installed as a Go tool rather than a system dependency:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

### Clipboard access

Quick Translate reads the text to translate from the clipboard and writes the translation back to it. How this is done depends on the operating system.

**On Linux**, the text is read from the *primary selection*, which holds the text currently marked with the mouse, because a copy shortcut such as `Ctrl+C+C` cannot be bound reliably on Linux. The translation is written to the regular clipboard instead, so it can be pasted with `Ctrl+V`. Reading and writing are both done through an external program, so **one** of the following packages is required at runtime:

- `wl-clipboard` >= 2.3.0, recommended for Wayland sessions
- `xclip` >= 0.13, for X11 sessions and also usable under XWayland
- `xsel` >= 1.2.1, as an alternative to `xclip`

The session type is determined on startup from the `XDG_SESSION_TYPE`, `WAYLAND_DISPLAY` and `DISPLAY` environment variables, and the first installed program that fits the session is used. The programs of the other session type are kept as a fallback. If none of the three packages is installed, Quick Translate reports an error on startup.

**On all other operating systems**, the regular clipboard is used for reading and writing, which is filled by the user pressing `Ctrl+C+C`. No extra packages are needed to be installed.

### Building

From the project root, run:

```sh
make build
```

This builds the frontend, compiles the application with Wails, and produces the binary at `build/bin/quick-translate`. Pass `UPX=0` to skip the compression while developing.

### Installing

To build the application, install the binary and register it with the desktop in one step, run:

```sh
make install
```

This chains the following targets, which can also be run individually:

- `make build` compiles the application.
- `make install-binary` copies the binary to `~/.local/bin`. It does not build, so a binary from a release can be installed the same way. Set `PREFIX` to install somewhere else.
- `make install-desktop` runs `quick-translate --install`, which registers the installed binary with the desktop: the application icon in eight sizes, a desktop entry, and a systemd user service that starts the application with the graphical session. Where no systemd user manager answers, an autostart entry is written instead.

The desktop environment is detected from the session. Pass `DESKTOP=kde` or `DESKTOP=generic` to choose it by hand:

```sh
make install DESKTOP=kde
```

On **KDE Plasma**, the desktop entry carries a global shortcut, so the application is bound to `Meta+T` right away, and window rules are merged into `~/.config/kwinrulesrc` so the frameless window stays on top and is kept out of the task bar and the window switcher. The shortcut can be changed under *System Settings > Keyboard > Shortcuts*.

On **every other desktop**, there is no portable way to bind a global shortcut, so this step is left to you: bind `~/.local/bin/quick-translate` to a key in the keyboard settings of your desktop. Everything else is installed the same way.

To see where the application has installed itself and whether it is running, use:

```sh
make status
```

### Uninstalling

To remove the desktop entry, the icons, the autostart, the window rules, and the binary itself, run:

```sh
make uninstall
```

**Your own files are kept.** Uninstalling never touches them, because reinstalling or rebuilding is far more common than leaving for good, and a history that disappeared with a rebuild would be a nasty surprise. What stays behind:

- `~/.config/quick-translate/` — the configuration you wrote, including your provider auth keys.
- `~/.local/share/quick-translate/` — the history of everything you translated.
- `~/.cache/quick-translate/` — the cached language lists, which are fetched again when they are missing.

To delete those as well, ask for it explicitly:

```sh
make uninstall PURGE=1
```

The binary understands the same as a flag, which is useful when the sources are no longer around:

```sh
quick-translate --uninstall --purge
```

> [!CAUTION]
> Purging cannot be undone. Your translation history and your provider configuration, auth keys included, are deleted for good.

## Provider

### DeepL

The DeepL provider allows using the DeepL API for translations. Both the free and paid API versions are supported. To use DeepL, the provider must be configured first. See the following sections on retrieving the auth key and configuring the provider for more information. DeepL supports a variety of languages and auto detection of the source language, making it ideal for setting it as the default provider.

#### Retrieving the DeepL Auth Key

The DeepL API requires a key for authentication, which can be retrieved from the official website. This works for both the free and paid version.

1. Visit the [official DeepL website](https://deepl.com), create an account or login.
2. Once logged in, click on `API plans` in the left sidebar and make sure you sign up for the preferred API plan. The free plan should be suitable for most users.
3. Afterward, visit your account settings via `Account` in the left sidebar and select the `API Keys & Limits` section.
4. Click on `Create key` and fill out the required details. Quick Translate uses only the `Translate text` and `Retrieve languages and resources` permissions. It is recommended to select only the required permissions for security reasons.
5. After filling out the form, the website will show you your auth key. Copy the key as it needs to be set in the configuration for Quick Translate to authenticate in your name to the DeepL API. You can retrieve the auth key later on the same website again.
6. Finally, paste the auth key into the provider configuration. See the configuration section for more information.

#### Configuration

Configuring the DeepL provider requires editing `config.yml` in your user configuration directory (`~/.config/quick-translate/config.yml`, or under `$XDG_CONFIG_HOME` when that is set). The file is created with a commented template the first time Quick Translate starts. Under the `provider` section specify the `deepl` key. The following options are supported and may be specified under the `deepl` key:

- `auth_key` (required): Set your personal auth key retrieved from the DeepL website. See the section above for more information.
- `free_version` (required): When using the free version of the API set this option to `true`. For the paid version use `false`.
- `fast_mode` (optional): DeepL provides latency-optimized and quality-optimized language models for translation. By setting `fast_mode` to `true`, the latency-optimized models are being used. Defaults to `false`.
- `formality` (optional): DeepL supports different formality styles per language. Not all languages are supported for this feature (see the [DeepL API documentation](https://developers.deepl.com/api-reference/translate#param-formality) for more information). The option supports `formal`, `informal` and `default`. Defaults to `default`.

## Contributing

Contributions are welcome, and so are bug reports — this is an alpha and it has only ever been run on a handful of machines. Please read [CONTRIBUTING.md](CONTRIBUTING.md) first: it covers how to build and verify a change, the conventions the codebase follows, and the one hard rule for pull requests (Conventional Commits in the title). Note that there is no test suite yet.

Questions of any kind are welcome as an [issue](https://github.com/manuelschoene/quick-translate/issues) or by email at [schoene-manuel@gmx.de](mailto:schoene-manuel@gmx.de).

Found a security problem? [SECURITY.md](SECURITY.md) explains how to report it, and also describes what the application does with your text, your auth keys and your history.

Everyone taking part is expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

Quick Translate is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more information.
