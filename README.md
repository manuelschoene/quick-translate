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

The DeepL provider allows using the DeepL API for translations. Both the free and paid API versions are supported. To use DeepL, the provider must be configured first. See the following sections on retrieving the auth key and configuring the provider for more information. DeepL supports a variety of languages and auto detection of the source language, making it ideal for setting it as the default provider.

#### Retrieving the DeepL Auth Key

The DeepL API requires a key for authentication, which can be retrieved from the official website. This works for both the free and paid version.

1. Visit the [official DeepL website](https://deepl.com), create an account or login.
2. Once logged in, click on `API plans` in the left sidebar and make sure, you sign up for the prefered API plan. The free plan should be suitable for most users.
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

## Dependencies

Default requirements:

- `upx` >= 5.2.0

When running Linux with Wayland the following packge is required:

`wl-clipboard` >= 2.3.0
`libnotify` >= 0.8.8

## License

The quick-translate project is licensed under the GPL-3.0-or-later. See the LICENSE file for more information.
