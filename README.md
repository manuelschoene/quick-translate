<p align="center">
	<img alt="Quick Translate Logo" src="art/quick-translate.png" width="256">
</p>

<br>

<p align="center">
	<a href="https://www.gnu.org/licenses/gpl-3.0.html">
		<img alt="License" src="https://img.shields.io/badge/License-GPL--3.0--or--later-blue.svg">
	</a>
	<a href="https://go.dev">
		<img alt="Go" src="https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white">
	</a>
	<a href="https://vuejs.org">
		<img alt="Vue.js" src="https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js">
	</a>
</p>

# Quick Translate

Translate focused text anywhere with a single shortcut.

## Provider

### DeepL

DeepL has two different APIs that allow translations. Currently only the free API is implemented. For the free API set the environment variable `DEEPL_AUTH_KEY` to the authorization token retrieved from the DeepL dashboard.

## Dependencies

By default, no external dependencies are required.

When running Linux with Wayland the following packge is required:

`wl-clipboard` >= 2.3.0
`libnotify` >= 0.8.8

## License

The quick-translate project is licensed under the GPL-3.0-or-later. See the LICENSE file for more information.
