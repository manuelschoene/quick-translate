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
