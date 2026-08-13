<p align="center">
	<img alt="Quick Translate Logo" src="art/quick-translate.png" width="256">
</p>

<br>

<p align="center">
	<a href="https://www.apache.org/licenses/LICENSE-2.0">
		<img alt="License" src="https://img.shields.io/badge/License-Apache--2.0-blue.svg">
	</a>
	<a href="https://go.dev">
		<img alt="Go" src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white">
	</a>
	<a href="https://vuejs.org">
		<img alt="Vue.js" src="https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js">
	</a>
</p>

# Quick Translate

> [!WARNING]
> Quick Translate is under active development. Features, configuration, and behavior may change without notice, and the project is not yet ready for production use.

Quick Translate is a lightweight desktop utility that translates clipboard or selected text on demand, triggered by a single global keyboard shortcut. Translation is performed by a configurable provider, described below.

Quick Translate is currently built and tested for Linux with KDE Plasma. Other desktop environments and operating systems are not yet fully supported.

## Installation

Installation is managed through the provided `Makefile`, which builds the application with Wails and installs the binary, a desktop entry, and an optional systemd user service. All targets are written for Linux.

### Prerequisites

- `Go` >= 1.26
- `Bun` >= 1.3.14, used to install and build the frontend
- `gcc` >= 16.1.1, required to build the CGO-based WebKit bindings used by Wails
- `pkgconf` >= 2.5.1, used by Wails to locate the GTK and WebKit libraries
- `webkit2gtk-4.1` >= 2.52.5, used by Wails to render the frontend on Linux (corresponds to the `webkit2_41` build tag used in the `Makefile`)
- `gtk3` >= 3.24.52, used by Wails on Linux
- `upx` >= 5.2.0, used to compress the built binary

The Wails CLI itself is not listed above, as it is installed as a Go tool rather than a system dependency:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
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

This builds the frontend, compiles the application with Wails, and produces the binary at `build/bin/quick-translate`.

### Installing

To build and install the binary, the systemd user service, and the desktop entry in one step, run:

```sh
make install
```

This chains the following targets, which can also be run individually:

- `make install-binary` builds the application and copies the binary to `~/.local/bin`.
- `make install-service` installs `quick-translate.service` to `~/.config/systemd/user` and enables it via `systemctl --user`.
- `make install-desktop` installs the application icon and a `.desktop` entry to `~/.local/share/applications`.

### Uninstalling

To remove the binary, the systemd service, and the desktop entry, run:

```sh
make uninstall
```

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

Configuring the DeepL provider requires editing the `provider.yaml` in your user configuration directory (e.g., `~/.config/quick-translate/provider.yaml`). Under the provider section specify the `deepl` key. The following options are supported and may be specified under the `deepl` key:

- `auth_key` (required): Set your personal auth key retrieved from the DeepL website. See the section above for more information.
- `free_version` (required): When using the free version of the API set this option to `true`. For the paid version use `false`.
- `fast_mode` (optional): DeepL provides latency-optimized and quality-optimized language models for translation. By setting `fast_mode` to `true`, the latency-optimized models are being used. Defaults to `false`.
- `formality` (optional): DeepL supports different formality styles per language. Not all languages are supported for this feature (see the [DeepL API documentation](https://developers.deepl.com/api-reference/translate#param-formality) for more information). The option supports `formal`, `informal` and `default`. Defaults to `default`.

## License

Quick Translate is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more information.
