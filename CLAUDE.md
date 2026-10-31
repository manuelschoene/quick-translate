# CLAUDE.md

Guidance for Claude Code in this repository. It states the rules and the traps; the details live in the files it
points to. Read those before changing them.

## What this is

A Wails v3 desktop app (Go + Vue 3) that translates the selected text (Linux primary selection) on a global
shortcut. It runs hidden, started with the session through an XDG autostart entry it registers itself. The
shortcut is bound by the running process, so it only works while the app runs. A second start reaches the first
through `SingleInstanceOptions` and shows the latest translation.

## Commands

[Task](https://taskfile.dev) from the repo root (`task --list` for all; `Taskfile.yml` → `build/Taskfile.yml`,
`build/linux/Taskfile.yml`), bun in `frontend/`:

```sh
task dev                        # live reload
task build [DEV=true|UPX=false] # -> bin/quick-translate
task archive                    # -> bin/quick-translate-<version>-linux-amd64.tar.gz
task common:generate:bindings   # -> frontend/bindings/   (tracked, CI checks it)
task common:generate:db         # -> internal/store/db/   (tracked, CI checks it with sqlc diff)
go vet ./...
bun run types | lint | format | build   # types is the only type check, build does not run it
```

There are no tests. Verify by running the app.

Traps:

- **Bindings:** regenerate after changing an exported method of a registered service, a type it returns, or an
  event registered with `application.RegisterEvent`. The task runs `wails3 generate bindings -ts` without `-b`,
  because `-b` imports the runtime from `/wails/runtime.js`, which `vue-tsc` cannot resolve.
- **Versions pinned in pairs:** `@wailsio/runtime` in `package.json` is the exact Wails version in `go.mod`.
  sqlc is pinned in `.github/actions/setup-sqlc` (bare version, no `v`), because every generated file names it.
- **Event types:** `bindings/github.com/wailsapp/wails/v3/internal/eventdata.d.ts` types `Events.CustomEvents`
  (hence `bindings/**/*.d.ts` in `tsconfig.json`); `eventcreate.ts` is loaded by the Wails Vite plugin, not by an
  import.
- **Builds need GTK ≥ 4.10 and WebKitGTK 6.0 dev packages**, also for `go vet`: Wails compiles its Linux part
  through cgo without a build tag. The version is stamped from `build/config.yml`.
- `lint:fix` in `package.json` ends in `|| true` for CI; never point a human at it.

## CI and releases

Branch from `main`, merge through a PR; CI only runs on `feature/`, `hotfix/`, `chore/`, `docs/`, `refactor/`,
`test/`. What it checks is in `.github/workflows/ci.yml`, the release flow in `.github/workflows/release.yml`
(`workflow_dispatch -f tag=…` rebuilds the assets of an existing release).

Versions come from `release-please-config.json`, never by hand:

- The `v` prefix is only on the tag and the release name. Everything that is data (`build/config.yml`, asset
  names, `INSTALL.md`) uses the bare version. Never name an asset after `tag_name`.
- `prerelease`, `versioning: prerelease` and `prerelease-type` **must stay together**. Without
  `prerelease: true`, `0.1.0-alpha` would be released as `0.1.0`.
- Only the counter after the label moves. Reach a new minor or stable version with a `Release-As: x.y.z`
  footer on `main`.

## Backend

Imports (`go list -f '{{.Imports}}'`), always downward:

```
translation → language, provider, store, models
language    → config, provider, store, models
provider    → models
store, config, clipboard, cli → system
main.go     → every service, wires them
```

Calls reach further than imports: `translation` uses `config` and `clipboard`, `provider` uses `config`, and
`language` uses `system`, each through an interface the consumer declares. The full picture with types is
`architecture.mmd`. Each package is one Wails service in
`internal/<domain>/service.go`; `main.go` (`wireServices`) only constructs and registers them.

| Package | Owns | Notes |
|---|---|---|
| `models` | `ProviderID` (+ slug), `Language`, `Translation` | no dependencies |
| `system` | `Files`, `Notifier`, desktop integration, locale, logger | bottom of the tree |
| `store` | SQLite connection, migrations, repositories | `internal/store/migrations`, `queries` → sqlc |
| `config` | the YAML file, read once at startup | template in `internal/config/file.go` |
| `provider` | the registry, the current `*Translator` | one file per provider |
| `language` | catalog per provider, selection, background refresh | pure rules in `selection.go` |
| `clipboard` | the backend (external program or library) | |
| `translation` | current translation, history and its cursor | lock order in `service.go` |
| `cli` | subcommands, the only terminal output | table in `internal/cli/command.go` |

Rules:

- **Dependencies point down only.** A package that reports to a higher one gets a hook at construction
  (`provider.OnChange`) or a struct of functions (`system.Actions`); it never imports it. Consumers declare the
  interfaces they need.
- **Registration:** a type is registered only if the frontend calls it or it needs `ServiceStartup`/
  `ServiceShutdown`. Every exported method the frontend must not reach carries `//wails:ignore`. Wrap a
  third-party service whose methods must not be bound (see `system.Notifier`). `system.Files` is never
  registered.
- **Order:** services start in registration order and shut down in reverse. The notifier comes first, so it can
  still report a failing startup.
- **A non-nil `ServiceStartup` error shuts the app down.** Only `store`, `config` and `clipboard` may return one,
  and they call `notifier.StartupFailed(err)` first. Everything else logs and degrades.
- **Events are the only way state reaches the frontend.** Mutations return only an error, getters serve the
  first load. Emit while holding the state's lock. Whatever a later event is resolved against is emitted
  first (`languages:changed` before `selection:changed`). Every `translation:started` gets exactly one
  `:changed` or `:failed`.
- **Never call a service inside `InvokeSync`/`InvokeAsync`:** an emitter waiting on a full queue would deadlock
  against the UI thread.
- **Adding a provider:** a `models.ProviderID` with its slug, a file implementing `provider.implementation` with
  a settings struct validated in its build function, and a `registry` row.
- **Adding a CLI command:** one row in `internal/cli/command.go` plus a file with its function. The CLI runs
  before `application.New`, because the single-instance lock is taken there.

## Runtime files

Every path, permission and purge rule is a row in `internal/system/files.go` (`files_linux.go` adds the
desktop integration). **A new or moved file is a row there and nowhere else.** Packages are handed a
`*system.Files` and ask for a resource (`files.Path(system.Settings)`); they never build paths or modes.
Directories are `0700`, files `0600`. The row's `category` decides what `status` reports and `purge` removes.

Non-obvious ones:

- **`~/.config/Quick Translate/`** (`NotificationRoot`): Wails' notifier writes its categories file there on
  every shutdown, with `0755`/`0644`. `system.Notifier` creates it first, and Wails keeps existing modes.
- **Desktop entry and icon** are installed from the release archive and only read. Both are named after
  `system.ApplicationID`, because the portal grants the shortcut to that id. GTK4 ignores `Options.Icon`, so the
  entry's `Icon=` is the only window icon.
- **Autostart** uses `ApplicationID` as identifier, not Wails' slug; the portal derives the app's identity from
  the unit named after that file.
- **KWin rules** are merged into `~/.config/kwinrulesrc` (`kde_linux.go`), written only when they differ.
- **`art/tray.png`** is embedded in `main.go`, because `//go:embed` cannot reach `art/` from `internal/system`.
  `task art` renders it.

## Frontend

Vue 3 `<script setup>`, Tailwind v4, vue-router with hash history, no Pinia.

```
views / components  →  composables/use*.ts  →  state.ts · router.ts · lib/ · services/ · @bind · @events
```

- **Components only import composables.** Each composable owns one state group in `state.ts`. It hands out
  read-only refs (`readonlyRefs`), computeds and actions, and exports `listen()` and `load()` for
  `useApplication.start()`.
- **Never read state right after `await`:** a promise resolves before the events of its own call.
- **Mutations compose the retranslation:** `request(async () => { await SetSource(tag); await retranslate(); })`.
- **Event names come from `@events`** (generated), listened to through `on()` in `lib/events.ts`, which only
  accepts declared names.
- **`services/request.ts` reports instead of throwing**: `request()` shows the error view, `requestQuietly()`
  does not.
- **Hash history**, because the asset server reserves `/wails/*` and has no `index.html` fallback.
- Aliases are kept in sync in `vite.config.ts` **and** `tsconfig.json`.

## Conventions

- Go: a prose doc comment above every function, unexported ones included, on behavior and edge cases. Same style
  in TSDoc.
- Errors: lowercase, no trailing punctuation, prefixed with the package, wrapped with `%w`
  (`config: decode file: …`).
- Logging goes through the `*slog.Logger` handed in (`system.NewLogger()`, stderr). Only `cli` prints.
- `import type` for type-only imports: `verbatimModuleSyntax` lets a value-style type import break the whole
  module graph at runtime.
- Import order is ESLint's job (`perfectionist`, `unused-imports`). Never add an import-sorting Prettier plugin.
- Run `bun run format`; never hand-sort attributes or Tailwind classes.

## Documentation is part of the change

A change is not finished until the docs describing it are true again. Re-read the sentence against the code;
do not trust that it was right.

| Changing … | Check |
|---|---|
| CLI commands or flags | `README.md`, `build/linux/INSTALL.md`, `internal/cli/command.go` |
| Tasks | `README.md`, `CONTRIBUTING.md`, the task's `desc` |
| Files, paths, permissions | `internal/system/files.go`, `README.md`, `SECURITY.md` |
| Configuration keys | `README.md`, `fileTemplate` in `internal/config/file.go` |
| What is sent or stored | `SECURITY.md` |
| Tool or dependency versions | `README.md`, `.github/actions/`, the traps above |
| Backend types or packages | `architecture.mmd`, "Backend" above |
| Conventions, workflow | `CONTRIBUTING.md`, "Conventions" above |

- `README.md`, `CONTRIBUTING.md` and `SECURITY.md` are public: no planning notes, no links to `PLANS.md`.
- Match each file's style: README paragraphs are one line each; `CLAUDE.md`, `CONTRIBUTING.md`, `SECURITY.md`
  and `INSTALL.md` wrap at 120 columns (`INSTALL.md` at 100).
