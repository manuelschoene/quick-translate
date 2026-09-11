# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Quick Translate is a Wails v2 desktop app (Go backend + Vue 3 frontend) that translates the text
currently selected (Linux primary selection) or copied, triggered by a global shortcut bound at the
desktop-environment level. It runs hidden as a systemd user service, or as an XDG autostart entry where no
systemd user manager answers; pressing the shortcut launches a *second* process which hands the request to
the running one over a unix socket and exits immediately.

## Commands

Backend / whole app (from the repo root). The Makefile picks the WebKit build tag with `pkg-config` —
`webkit2_41` when webkit2gtk-4.1 is installed and none when only 4.0 is — so the tag is never written by
hand:

```sh
make dev                     # wails dev with the detected tag
make build                   # -> build/bin/quick-translate; UPX=0 skips the slow compression
make install                 # build + copy the binary + 'quick-translate --install'
make install DESKTOP=kde     # force the Plasma flavour instead of detecting it from the session
make status                  # what the installation looks like on this machine
make uninstall               # remove the desktop integration and the binary; user files are kept
make uninstall PURGE=1       # ... and delete config, history and cache as well (irreversible)
make icons                   # re-render the checked-in icon set from the master artwork
go vet ./...
```

The version is stamped in from `wails.json`'s `info.productVersion` via `-ldflags -X main.version=…`; a
build made on an exact git tag uses the tag instead. A build without the flag reports `0.0.0-dev`.

Building and installing are deliberately separate targets: `make install-binary` works on a binary that
was downloaded rather than built, so a release needs no toolchain.

Work happens on `feature/…` or `hotfix/…` branches cut from `main` and lands through a pull request.
`main` is protected and is the nightly state people clone and build, so never commit to it directly and
never leave it in a state that does not install. CI (`.github/workflows/ci.yml`) runs one job that
auto-formats and commits the result back, and one that checks: gofmt, `go vet`, `go test`, `go mod verify`,
a `go mod tidy` that must change nothing, `govulncheck`, the frontend checks, a full `make build`, and
`desktop-file-validate` over the entry the binary generates.

`.github/workflows/release.yml` runs when a release is published: it calls the CI workflow, then builds
twice on `ubuntu-22.04` — once with `WEBKIT_TAGS=webkit2_41` and once with `WEBKIT_TAGS=` — and attaches
both archives plus a `SHA256SUMS` file. The old runner is for glibc, not WebKit: glibc is not forward
compatible, so a binary built on 24.04 will not start on Debian 12. It also refuses to build when the tag
disagrees with `wails.json`'s `productVersion`. Note that the binary Wails produces has no section headers,
so neither `ldd` nor `strings` can tell you which WebKitGTK it links — the pinned `WEBKIT_TAGS` and a build
that fails on a missing pkg-config package are the only guarantee there is.

Frontend (from `frontend/`, package manager is **bun**, not npm):

```sh
bun install
bun run types    # vue-tsc --noEmit — the only type check; also runs as part of `bun run build`
bun run lint     # eslint --fix
bun run format   # prettier --write --no-editorconfig
bun run build    # types + vite build
```

`lint` and `format` are the developer-facing commands. CI calls the `lint:check`/`lint:fix` and
`format:check`/`format:fix` variants instead, which add a cache location the workflow restores between
runs — `lint:fix` also carries `--quiet || true`, so never point a human at it.

There are no tests in this repository yet (no `*_test.go`, no frontend test runner). Verify changes by
running the app.

`frontend/wailsjs/` is generated but **tracked**, so a fresh checkout type-checks and lints without a Go
toolchain. After adding, removing, or changing an **exported** method on `transport.Adapter` or a DTO,
regenerate the bindings and commit them together with the change; stale bindings make the frontend fail at
a distance, and nothing else catches it.

Only `wails build` and `wails dev` regenerate them. `wails generate module` does **not** — it runs, prints
nothing and writes no files. And `wails build` cannot regenerate them while the application is running:
the generation step builds the app and executes it, `transport.Connect()` hands over to the instance that
already holds the socket and exits before `wails.Run` is reached, so the step reports "Generating bindings:
Done." without writing anything and the build then fails on the missing modules. Stop the service first:

```sh
systemctl --user stop quick-translate.service
make build
systemctl --user start quick-translate.service
```

## Runtime files

- Config: `~/.config/quick-translate/config.yml` — created with a commented default template on first
  start (`internal/config/file.go`), mode `0600` in a `0700` directory because it holds the provider keys.
  Existing installs are narrowed on start by `narrow()` in the same file.
- History DB (SQLite, pure-Go driver): `~/.local/share/quick-translate/history.db` (`$XDG_DATA_HOME`) —
  data the user produced and cannot re-fetch, so neither next to the hand-written config nor in the cache,
  which cleanup tools may empty. Directory `0700`, database `0600`; it holds every translated text.
- Language cache (gob, 7-day TTL, one file per provider): `~/.cache/quick-translate/translations/<slug>.gob`.
- IPC socket: `$XDG_RUNTIME_DIR/quick-translate.sock` (falls back to the temp dir).

Written by `quick-translate --install`, all resolved through the freedesktop base directories, so
`XDG_DATA_HOME` and `XDG_CONFIG_HOME` are honoured:

- Desktop entry: `~/.local/share/applications/quick-translate.desktop`. The KDE flavour is the same entry
  plus `X-KDE-Shortcuts`, installed under the same name so switching flavours replaces it.
- Icon: `~/.local/share/icons/hicolor/<size>/apps/quicktranslate.png` in eight sizes (16 … 256),
  referenced by name from the entry. The name carries no dash on purpose: an icon name the current theme
  does not know is retried with everything after its last dash cut off, so `quick-translate` resolves to
  Breeze's unrelated `quick` icon before hicolor is ever reached. Verify a change with `kiconfinder6
  quicktranslate`, which is the loader Plasma itself uses.
- Autostart: `~/.config/systemd/user/quick-translate.service`, or `~/.config/autostart/quick-translate.desktop`
  when no systemd user manager answers.
- Window rules (KDE only): merged into `~/.config/kwinrulesrc`, which has no drop-in directory.

## Backend architecture

Strict one-direction layering; each layer only knows the one below it:

```
main.go → transport (Adapter) → core (Core) → config · clipboard · history · language · provider → models
main.go → desktop
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
  `ipc.go` implements the single-instance socket; `Connect()` in `main.go` returns true when another
  instance took the request over.
- **desktop** — the installation, a side branch that only `main.go` reaches and that knows nothing about
  the rest of the app. `Install`/`Uninstall`/`Status` are the whole API; the per-OS work sits behind build
  tags the way `clipboard` and `transport` do it, with `linux.go` (icon, entry, systemd or XDG autostart),
  `kde.go` (the KWin rule merge) and `unsupported.go` for the operating systems that follow later. The
  files it writes are `text/template`s embedded from `internal/desktop/assets/linux/`, rendered with the
  path of the *running* binary, so the install never copies or moves anything itself. The icon set beside
  them (`assets/linux/icons/<size>.png`) is checked in rather than scaled at runtime; `make icons`
  re-renders it from `art/quick-translate.png` and is the only target that needs ImageMagick. `Purge()` is
  the one place that knows all three per-user directories at once (`<config|data|cache>/quick-translate`);
  if `config`, `history` or `language` ever moves its files, `paths` in `linux.go` has to follow.

Two paths reach the frontend, and both must stay supported:

1. **Frontend-initiated** — a bound `Adapter` method returns a DTO (or a Go error, which Wails rejects
   the promise with as a plain string).
2. **Shortcut-initiated** — the socket wakes `Adapter.show()`, which emits the events `translating`,
   then `translation` or `error`. These names are duplicated as constants in
   `internal/transport/adapter.go` and `frontend/src/services/events.ts`; change both together.

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
 services/events.ts · wailsjs bindings
```

The split is vertical by domain in `composables/` and horizontal by role below it. The one seam that
has to stay central is `services/wire.ts`: a `TranslationDto` carries the chosen languages, the
translated text and the history flags at once, so whoever unpacks it reaches into three state groups
anyway.

- `src/state.ts` — the `reactive` state groups, the only mutable state in the app, plus the
  `Language` type they hold. Written by `services/`, read through the composables.
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
- `services/events.ts` — the three backend events, mirrored from `internal/transport/adapter.go`.
- `lib/` — pure helpers with no app state: `cn`, `keyboard` (`onKey`), `search`, and `reactivity`
  (`readonlyRefs`, which wraps a state group in `readonly` + `toRefs` so components can only read).
- Types live with what owns them: `Language` in `state.ts`, `ViewName` in `router.ts`, `ButtonProps`
  in a plain `<script>` block of `Button.vue`. There is no types-only module.
- A component that is styled from the outside declares a `class` prop and merges it with `cn()`
  (`Button`, `Panel`, `LoadingTextarea`, and the toolbars, which forward it down to `Panel`); one that
  is never styled from the outside declares none. Templates address props as `props.x`, never bare.
- Import aliases (kept in sync in `vite.config.ts` **and** `tsconfig.json`): `@`, `@assets`, `@comp`,
  `@lay`, `@lib`, `@services`, `@use`, `@views`, `@wails`.

## Conventions

- Go: a full-sentence prose doc comment above every function, including unexported ones, describing
  behaviour and edge cases rather than restating the signature. Frontend files use the same prose style
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
behaviour a user or a contributor ever sees, and with no test suite there is nothing that contradicts them
when they drift. **A change is not finished until the files that describe it are true again.**

Before calling any change done, walk this list and update what the change touched:

| Changing … | Check |
|---|---|
| CLI commands or flags | `README.md`, `CONTRIBUTING.md`, the `usage` string in `main.go` |
| `make` targets or variables | `README.md`, `CONTRIBUTING.md`, the `help` target in the `Makefile` |
| Where files are written, or their permissions | `README.md` (uninstalling), `SECURITY.md`, "Runtime files" above |
| Configuration keys | `README.md`, the `defaultYml` template in `internal/config/file.go` |
| What the app sends over the network, or stores | `SECURITY.md` |
| Prerequisites, dependency or tool versions | `README.md`, `CONTRIBUTING.md`, "Commands" above |
| Backend types or package boundaries | `architecture.mmd`, "Backend architecture" above |
| Conventions, tooling, workflow | `CONTRIBUTING.md`, "Conventions" above |
| Runtime paths, permissions, layering | this file |

Three rules that apply to all of them:

- **Verify the claim, do not assume it.** Documentation rots quietly. This repository has already shipped a
  README that told users to edit `provider.yaml` when the code created `config.yml`, and that pinned a Wails
  version two minor releases behind. Both survived because nobody re-read them against the code. When you
  touch a documented behaviour, open the file and check the sentence, do not trust that it was right.
- **`README.md`, `CONTRIBUTING.md` and `SECURITY.md` are for users and contributors.** Internal planning
  notes and review documents stay out of them; do not link the working documents in the repository root
  from the public-facing files.
- **Match the file's own style.** README paragraphs and list items are each on a single line, however long;
  `CLAUDE.md`, `CONTRIBUTING.md` and `SECURITY.md` wrap at 120 columns. Rewrapping a file you only edited
  one sentence in makes the diff unreadable.

