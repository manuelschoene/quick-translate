# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Quick Translate is a Wails v3 desktop app (Go backend + Vue 3 frontend) that translates the text
currently selected (Linux primary selection) or copied, triggered by a global shortcut the running process
binds itself through Wails. It runs hidden, started with the session through an XDG autostart entry the
application registers itself; the shortcut press arrives as a callback in that process, so no second process
is started and the shortcut only works while the application runs. Starting the binary a second time reaches
the first one through `SingleInstanceOptions` and shows the last translation instead of making a new one.

## Commands

Build system is [Task](https://taskfile.dev): `Taskfile.yml` delegates to `build/Taskfile.yml` (shared) and
`build/linux/Taskfile.yml`. From the repo root:

```sh
task dev                        # wails3 dev with live reload
task build                      # -> bin/quick-translate, production tags, UPX-compressed
task build DEV=true             # no production tag, no compression, version stays 0.0.0-dev
task build UPX=false            # production build without compression (what CI uses)
task archive                    # -> bin/quick-translate-<version>-linux-amd64.tar.gz
task art                        # re-render art/tray.png and art/wordmark.svg (Inkscape + Noto Sans)
task common:generate:bindings   # regenerate frontend/bindings/
go vet ./...
```

Only the GTK4 / WebKitGTK 6.0 build is supported; Wails calls GTK 4.10 APIs, so the build needs GTK ≥ 4.10.
The version comes from `info.version` in `build/config.yml`, stamped in with `-ldflags -X main.version=…`.
`generate:bindings`, `install:frontend:deps` and `build:frontend` carry fixed `label`s, so their checksums in
`.task/` are shared between the `common:` and `linux:common:` call paths.

The archive mirrors `~/.local`: `bin/`, `share/applications/<ApplicationID>.desktop`,
`share/icons/hicolor/scalable/apps/<ApplicationID>.svg`, plus `INSTALL.md` rendered from
`build/linux/INSTALL.md`. The `render` task fails on any `@PLACEHOLDER@` it does not fill.

Frontend (from `frontend/`, package manager is **bun**, not npm):

```sh
bun run types    # vue-tsc --noEmit, the only type check; not part of `bun run build`
bun run lint     # eslint --fix
bun run format   # prettier --write --no-editorconfig
bun run build    # vite build only
```

CI calls the `lint:check`/`lint:fix` and `format:check`/`format:fix` variants, which add a cache location;
`lint:fix` carries `--quiet || true`, so never point a human at it.

There are no tests (no `*_test.go`, no frontend test runner). Verify changes by running the app.

`frontend/bindings/` is generated but **tracked**, so a checkout type-checks without a Go toolchain. After
changing an **exported** method on `transport.Adapter`, a DTO, or an event registered with
`application.RegisterEvent`, run `task common:generate:bindings` and commit the result; CI fails on stale
bindings. It runs `wails3 generate bindings -ts` without `-b`: `-b` imports the runtime from
`/wails/runtime.js`, which only the running app serves, so `vue-tsc` could resolve neither the runtime nor the
event types. `@wailsio/runtime` in `package.json` is pinned to the exact Go module version; bump both together.

Event types reach the frontend in two parts. `bindings/github.com/wailsapp/wails/v3/internal/eventdata.d.ts`
fills `Events.CustomEvents`, which is why `tsconfig.json` includes `bindings/**/*.d.ts`; without it every
event's `data` is silently `any`. `eventcreate.ts` turns the data into DTO classes at runtime and is loaded
by the `@wailsio/runtime/plugins/vite` plugin in `vite.config.ts`, not by any import. The runtime types are
parameterized on the event *name* (`Events.WailsEvent<'translation'>`), never on the data type.

## CI and releases

Work happens on branches from `main`, landing through a pull request. CI only triggers on `feature/`,
`hotfix/`, `chore/`, `docs/`, `refactor/` and `test/`. `main` is protected and must always install.

`.github/workflows/ci.yml` has two jobs: one auto-formats and commits back, the other checks (frontend checks,
gofmt, `go vet`, `go test`, `go mod verify`, unchanged `go mod tidy`, `govulncheck`, `task archive UPX=false`
with `desktop-file-validate`, current bindings, a CLI smoke test). Wails v3 compiles its Linux frontend
through cgo without a build tag, so every Go step needs the GTK4/WebKitGTK 6.0 dev packages.

`.github/actions/setup-wails` installs those packages, checks `pkg-config webkitgtk-6.0`, optionally
`desktop-file-utils` and UPX 5.2.1 (inputs `desktop-file-utils`, `upx`), and `wails3` at the `go.mod` version.

`.github/workflows/release.yml` runs on every push to `main`:

1. `release-please` keeps the release PR up to date, or, once it is merged, tags and creates the release.
   Only then is `release_created` set, which gates the rest.
2. `resolve` hands `tag` and `version` down. It is the only job that knows the trigger.
3. `verify` runs the CI workflow again.
4. `build` runs `task archive` on `ubuntu-24.04`, the oldest runner with GTK ≥ 4.10, after checking
   `build/config.yml` against the release version.
5. `publish` attaches the archive and a `SHA256SUMS` file.

`workflow_dispatch` with a `tag` input rebuilds the assets of an existing release
(`gh workflow run release.yml -f tag=v0.1.0-alpha`). That path skips `release-please` (it can never create or
move a release) and `verify` (an old tree can fail on newly published vulnerabilities), and `resolve` fails
early when no release exists. Because `verify` may be skipped, `build` needs
`!cancelled() && needs.resolve.result == 'success' && needs.verify.result != 'failure'`.

Versions come from `release-please-config.json`, never by hand:

- `initial-version` is the first release; otherwise release-please starts at `1.0.0`.
- `extra-files` (`generic` updater) rewrites the line marked `x-release-please-version` in
  `build/config.yml`, the single source the Taskfiles read.
- The `v` prefix is split: tag and release name keep it (`v0.1.0-alpha`), everything that is data (config,
  asset names, `INSTALL.md`, `--version`) reads `0.1.0-alpha`. The updater and the action's `version` output
  carry the bare semver, only `tag_name` has the prefix. Never name an asset after `tag_name`.
- `prerelease`, `versioning` and `prerelease-type` **must stay together**. `prerelease: true` makes the GitHub
  release a prerelease. **Never set `versioning: prerelease` without it:** that drops the label, so
  `0.1.0-alpha` would be released as `0.1.0`.
- With `versioning: prerelease` only the counter after the label moves (`0.1.0-alpha` → `0.1.0-alpha.1`),
  whatever the commit type. To reach `0.2.0-alpha` or a stable version, put `Release-As: 0.2.0-alpha` in the
  footer of a commit on `main`.
- `bump-minor-pre-major` keeps a breaking change inside `0.x` until 1.0.

## Runtime files

- Config: `~/.config/quick-translate/config.yml` (`system.Settings`) — created with a commented default
  template on first start (`internal/config/file.go`) because it holds the provider keys.
- History DB (SQLite, pure-Go driver): `~/.local/share/quick-translate/history.db` (`system.History`) —
  data the user produced and cannot re-fetch, so neither next to the hand-written config nor in the cache,
  which cleanup tools may empty. It holds every translated text.
- Language cache (gob, 7-day TTL, one file per provider): `~/.cache/quick-translate/translations/<slug>.gob`
  — one row, `system.Languages`, addressed with the file name as a part.

Every one of those directories is `0700` and every file inside them `0600`, and none of it is written in the
package that reads the file: the path, the permissions and the purge all come from the one table in
`internal/system/files.go`. A package is handed a `*system.FileService` and asks it for a resource by name
(`files.Path(system.Settings)`), never building a path or a mode of its own, so the layout of an installation
is stated once. `github.com/adrg/xdg` resolves the base directories there, never by reading `XDG_*` out of
the environment by hand.

Installed by whoever unpacks the release archive into `~/.local`, and only ever *read* by the binary, which
reports them under `--status`. Both are named `<ApplicationID>`, because the portal grants the global
shortcut to the id the desktop entry is named after:

- Desktop entry: `~/.local/share/applications/io.github.manuelschoene.QuickTranslate.desktop`, generated by
  `generate:dotdesktop` in `build/linux/Taskfile.yml` from `build/linux/app.desktop`.
- Icon: `~/.local/share/icons/hicolor/scalable/apps/io.github.manuelschoene.QuickTranslate.svg`, one SVG.
  GTK4 ignores `Options.Icon` and `LinuxWindow.Icon`, so the `Icon=` key of the entry is the only way a
  window gets an icon at all.

Written by the running application, and removed again by `--purge`:

- Autostart: `~/.config/autostart/io.github.manuelschoene.QuickTranslate.desktop`, written on every start by
  `Integration.ServiceStartup` through `app.Autostart`. The identifier is the application id and not
  Wails' derived slug, because the desktop turns the entry into a unit named after the file and the portal
  derives the application's identity from that unit. There is no systemd unit and no restart on failure any
  more.
- Window rules (KDE only): merged into `~/.config/kwinrulesrc`, which has no drop-in directory, by
  `ensureRules` on every start. It builds what the file would have to look like and writes only when that
  differs from what is there, so the usual start writes nothing and a rule the user broke is repaired.

## Backend architecture

Strict one-direction layering; each layer only knows the one below it:

```
main.go → transport (Adapter) → core (Core) → config · clipboard · history · language · provider → models
main.go → system
main.go → cli → system

config · history · language · core · transport · cli → system   (a *FileService, handed in)
```

- **models** — plain structs (`Translation`, `Language`, `LanguagePreferences`) and the `Provider`
  interface. Has no dependencies; every other package speaks in these types.
- **provider** — one file per translation service plus a registry: `All()` maps slug → "supports
  language detection". Adding a provider means: implement `models.Provider`, add the slug constant and
  the `All()` entry, and add a case to `buildProvider` in `internal/core/provider.go`. Provider structs
  carry their own `yaml` tags and are filled by `config.LoadProvider`.
- **language** — `Collection` holds the current source/target/detected tags plus the provider's language
  lists, built through `NewBuilder()`. Tags are BCP 47 and validated with `golang.org/x/text/language`.
  `LanguageDetectionTag` (`"auto"`) is a virtual tag: it is never part of the language lists, and a
  provider is asked to detect when the source tag is empty. Setting one language may switch or drop the
  other when they collide, so callers must read both back after a change.
- **core** — the single stateful owner (provider, collection, current translation, history) behind one
  `sync.Mutex`. Every exported method locks; unexported helpers are documented "Requires the lock to be
  held" and must be called with it. State swaps happen atomically (`c.providerSlug, c.provider, c.langs
  = …`) so a failed provider switch leaves the previous one fully usable. Repeating an identical
  text/languages/provider combination returns the cached translation rather than paying the API again.
- **transport** — `Adapter` is the *only* type bound into Wails; its exported methods are the frontend
  API and its DTOs (`TranslationDto`, `LanguageDto`, `ProviderDto`, `FullDto`) are the wire format.
  `Show()` and `Restore()` are exported for the global shortcut and the second-instance callback, which
  live outside the frontend, and carry `//wails:ignore` so the bindings generator leaves them out.
- **system** — where the application meets the operating system it runs in: every path it touches, and what
  it has to write to be found by the desktop. It knows nothing about the rest of the app, but the rest reads
  its paths, so it is the bottom of the tree rather than a side branch. `ApplicationID`, `Shortcut`,
  `TrayLabel` and `TrayTooltip` live here, and every other place that needs a name or the keys reads them
  from here. `IntegrationService` is handed an `Actions` struct of the functions the shortcut and the tray
  entries trigger rather than importing `transport`, which is what keeps the direction.
  `FileService` (`files.go`) is the one table of every file and directory of an installation: each row
  carries what the file is for, where it lives, the permissions it is created with, who put it there and the
  directory row it sits in. Its `Path`/`Exists`/`Read`/`Write`/`EnsureFile`/`EnsureDirectory`/`Narrow`/`Remove`
  are what every package uses instead of building a path, an `os.MkdirAll` or a `Chmod` of its own, and a row
  addressed with extra parts serves a directory of files such as the language cache. **A file that moves, or
  a new one, is a row in that table and nowhere else.**
  There is no package-level instance and no accessor: `wireServices` in `main.go` builds one with
  `system.NewFileService()` and gives it to the two Wails services, and the `Adapter` passes it into the
  core, which passes it into `config`, `history` and `language`. `cli.Run` builds one of its own,
  because the command line runs before the application exists and nothing can hand it one. It is not
  registered with Wails: registering it would make the bindings generator expose `Read`, `Write` and `Remove`
  to the frontend, which has no business reaching the file system.
  The `category` of a row is what makes the rest derive from it — `internal`, `data`, `registered`,
  `installed` and `merged`: `Removals`, `Purge` and `Status` walk the table rather than listing what they
  know, `Remove` refuses a row the release archive installed and the one file the user owns, and an
  `internal` row is only ever handed out as a path, which keeps the report at the granularity the user cares
  about. Every other category carries the `name` the user knows the file by, so the table says which rows are
  reported instead of leaving it to an empty field. `Purge` answers with what it removed and prints nothing;
  `cli` renders it.
  What differs per operating system is only the table, so `files.go` carries the type, every operation and
  the rows every operating system shares, while `entries()` is defined once per build tag: `linux.go` adds
  the desktop entry, the icon, the autostart and the KWin rule file, and `unsupported.go` adds nothing,
  which leaves a not-yet-supported operating system with working per-user paths and no integration.
  `integrate`/`Purge`/`Removals`/`Status` are likewise defined per operating system rather than wrapped — only
  the three the command line calls are exported, the integration is reached through its service — and
  `kde.go` holds the KWin rule merge, the only thing the application writes that cannot ship as a file in
  the release archive. `integration.go` is the Wails service that integrates, registers the autostart, binds
  the global shortcut and puts the application into the tray at startup, and the place a settings view will
  reach for to switch any of them; it is the one file of the package that knows Wails. None of its failures
  are returned: a non-nil return from `ServiceStartup` takes the whole application down.
  The tray is the only thing on screen while the application runs hidden. It is a StatusNotifierItem over
  the session bus on Linux, which Plasma shows natively and GNOME only with the AppIndicator extension, and
  Wails answers a missing host through the error handler rather than a failure. Its icon has to be a PNG —
  `art/tray.png`, written by `task art` from `art/icon.svg` and committed, because the binary embeds it.
  The embed lives in `main.go` and the bytes are handed in: `//go:embed` cannot reach out of its own
  directory, so a file under `art/` is out of reach from `internal/system`.
- **cli** — the commands the binary understands besides starting the application (`--status`, `--purge`,
  `--version`, `--help`), and the only place that renders anything for a terminal. `system` answers with
  data and prints nothing the CLI has to parse, so a command is one entry in the table in `command.go` plus
  the function it names, in a file of its own (`usage.go`, `status.go`, `purge.go`); the usage text is built
  from that same table. Runs before `application.New`, and has to: the single-instance lock is taken inside
  it, so a command reaching that far would be forwarded to the running instance and exit silently.
  A command reaches the user only through the `interaction` it is handed (`interaction.go`): it embeds
  `io.Writer`, so printing is unchanged, and adds `confirm(confirmation)` for a yes-or-no question. The
  only implementation is `terminal`, over the streams the binary was started with, and it is built once in
  `Run` — a second one would open a second buffered reader on the same input. A `confirmation` carries the
  question, the words before the `[y/N]` and what an empty answer falls back to; nothing but a single `y`
  or `n` counts, and an input that ended falls back rather than asking again, which is what keeps a command
  from hanging where there is no terminal.

Three paths reach the frontend, and all of them must stay supported:

1. **Frontend-initiated** — a bound `Adapter` method returns a DTO (or a Go error, which Wails rejects
   the promise with as a plain string).
2. **Shortcut-initiated** — the global shortcut calls `Adapter.Show()`, which emits the events
   `translating`, then `translation` or `error`.
3. **Second launch** — `OnSecondInstanceLaunch` calls `Adapter.Restore()`, which emits `restored` with a
   `FullDto`, because a stored translation brings its own provider and languages with it. An empty or
   disabled history is printed and swallowed; the window is open either way.

The four event names are duplicated as constants in `internal/transport/adapter.go` and
`frontend/src/services/events.ts`; change both together.

Mutation calls that change languages or the provider re-translate the current text before returning
(`retranslate()`), which is why they answer with state rather than nothing. Provider changes and history
steps return `FullDto`, because the selectable languages belong to the provider.

## Frontend architecture

Vue 3 `<script setup>` + Tailwind v4 (no vue-router, no Pinia). Data flows one way:

```
views / components          only ever talk to composables
        ↓
composables/                read-only refs + derived computeds + actions   (one per domain)
        ↓            ↓
    state.ts      lib/ · router.ts                                         (no domain knowledge)
        ↑
 services/wire.ts           the only module that knows Go field names
        ↑
 services/events.ts · bindings/
```

The split is vertical by domain in `composables/` and horizontal by role below it. The one seam that
has to stay central is `services/wire.ts`: a `TranslationDto` carries the chosen languages, the
translated text and the history flags at once, so whoever unpacks it reaches into three state groups
anyway.

- `src/state.ts` — the `reactive` state groups, the only mutable state in the app. Written by `services/`,
  read through the composables.
- `src/router.ts` — `ViewName`, the `ViewName → Component` table and `route()`. The single door to
  navigation; there is no vue-router. `component` is a `ref` the App renders with `<component :is>`.
- `composables/use*.ts` — the only thing views and components import. Each hands out read-only refs
  of its state group, the computeds derived from it (labels, sorted lists, resolved providers) and
  the actions that change it. The returned object is built once at module level and handed out as
  `typeof`, so it neither rebuilds per component nor needs its type written out.
- `services/request.ts` — wraps every backend call. `request()` counts into `requestState.pending`
  and sends a failure to the error view; `requestQuietly()` does neither and is for the calls beside
  the translation (copy, hide). Both **report errors instead of throwing**, so no caller above needs
  try/catch. The shortcut translation is a flag, not a count, because the frontend only sees its two
  ends as events.
- `services/report.ts` — `report()` notes a failure and shows the error view, `reportQuietly()` only
  notes it.
- `services/wire.ts` — the only place that knows Go field names; maps DTOs onto state.
- `services/events.ts` — the four backend events, mirrored from `internal/transport/adapter.go`.
- `lib/` — pure helpers with no app state: `cn`, `keyboard` (`onKey`), `search`, and `reactivity`
  (`readonlyRefs`, which wraps a state group in `readonly` + `toRefs` so components can only read).
- Types live with what owns them: `Language` and the DTOs in the generated `@bind` modules, `ViewName` in
  `router.ts`, `ButtonProps` in a plain `<script>` block of `Button.vue`. There is no types-only module.
- A component that is styled from the outside declares a `class` prop and merges it with `cn()`
  (`Button`, `Panel`, `LoadingTextarea`, and the toolbars, which forward it down to `Panel`); one that
  is never styled from the outside declares none. Templates address props as `props.x`, never bare.
- Import aliases (kept in sync in `vite.config.ts` **and** `tsconfig.json`): `@`, `@assets`, `@comp`,
  `@lay`, `@lib`, `@services`, `@use`, `@views`, `@bind`.

## Conventions

- Go: a full-sentence prose doc comment above every function, including unexported ones, describing
  behavior and edge cases rather than restating the signature. Frontend files use the same prose style
  in TSDoc blocks.
- Error strings are user-facing GUI text: capitalized, punctuated full sentences that usually tell the
  user what to do ("Please set 'history.max_entries' …"). Wrap with `%w` when the cause matters.
- Failures that only cost a feature (unreadable history, unresolvable detected language, unstorable
  translation) are printed to stdout and swallowed; only failures that make the app unusable are
  returned as errors.
- Prettier is configured with 4-space indent, single quotes, 120 columns, and plugins that organize
  attributes and Tailwind classes — run `bun run format` rather than hand-sorting.
- Import order is an ESLint concern, not a Prettier one: `perfectionist` sorts declarations
  (non-relative first, relative last, alphabetical within both) and named specifiers (values before
  types), `unused-imports` deletes what is no longer used. It runs on `bun run lint`, not on
  `bun run format`, and unlike the old `prettier-plugin-organize-imports` it also reaches the script
  blocks of `.vue` files. Never add an import-sorting Prettier plugin back — the two would fight.
- Frontend imports that only bring in a type must say so (`import type { … }`, or a `type` prefix on the
  single specifier of a mixed import). `verbatimModuleSyntax` is on, so Vite hands the import through to
  the browser untouched: a value-style import of a type resolves to a missing export at runtime and takes
  the whole module graph down. `bun run types` catches it.
- `architecture.mmd` is a Mermaid class diagram of the Go packages; update it when backend types or
  package boundaries change.

## Documentation is part of the change

The prose files in this repository are maintained code, not decoration. They are the only description of
behavior a user or a contributor ever sees, and with no test suite there is nothing that contradicts them
when they drift. **A change is not finished until the files that describe it are true again.**

Before calling any change done, walk this list and update what the change touched:

| Changing … | Check |
|---|---|
| CLI commands or flags | `README.md`, `CONTRIBUTING.md`, the command table in `internal/cli/command.go` |
| Tasks or task variables | `README.md`, `CONTRIBUTING.md`, the task's `desc`/`summary` in the Taskfile |
| Where files are written, or their permissions | `internal/system/files.go`, `README.md`, `SECURITY.md` |
| Configuration keys | `README.md`, the `defaultYml` template in `internal/config/file.go` |
| What the app sends over the network, or stores | `SECURITY.md` |
| Prerequisites, dependency or tool versions | `README.md`, `build/linux/INSTALL.md`, `.github/actions/`, "Commands" above |
| Backend types or package boundaries | `architecture.mmd`, "Backend architecture" above |
| Conventions, tooling, workflow | `CONTRIBUTING.md`, "Conventions" above |
| Runtime paths, permissions, layering | this file |

Three rules that apply to all of them:

- **Verify the claim, do not assume it.** Documentation rots quietly. This repository has already shipped a
  README that told users to edit `provider.yaml` when the code created `config.yml`, and that pinned a Wails
  version two minor releases behind. Both survived because nobody re-read them against the code. When you
  touch a documented behavior, open the file and check the sentence, do not trust that it was right.
- **`README.md`, `CONTRIBUTING.md` and `SECURITY.md` are for users and contributors.** Internal planning
  notes and review documents stay out of them; do not link the working documents in the repository root
  from the public-facing files.
- **Match the file's own style.** README paragraphs and list items are each on a single line, however long;
  `CLAUDE.md`, `CONTRIBUTING.md` and `SECURITY.md` wrap at 120 columns. Rewrapping a file you only edited
  one sentence in makes the diff unreadable.

