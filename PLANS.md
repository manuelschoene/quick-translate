# Part 1 — From the adapter to services

## Goal

`core.Core` owns every piece of state behind one mutex, and `transport.Adapter` answers every call with DTOs
that bundle all of it. Both go away. Each piece of state gets one owning service that talks to the frontend
through its own bindings and events. The packages underneath (the language matching, `clipboard`, the providers,
the history SQL) keep their logic; only their wiring and interfaces change.

## Services

Every service lives in `internal/<domain>/service.go` (bindings: `@bind/<domain>/service`).

| Service | Owns | Bound | Internal (`//wails:ignore`) |
|---|---|---|---|
| `config` | Access to the configuration | – | `DefaultProvider`, `Providers`, `Provider(ID)`, `Preferences`, `HistoryLimit` |
| `provider` | Catalog, current translator | `Providers()`, `Current()`, `Select(ID)` | `Translator()`, `Fetcher(ID)`, `OnChange(hook)` |
| `language` | The catalog (lists, preferences) and the selection | `Options()`, `Selection()`, `SetSource`, `Detect`, `SetTarget`, `Switch` | `Snapshot()`, `Adopt(entry)`, `SetDetected(ID, tag)` |
| `clipboard` | The backend (program or library), selected at startup | – | `Read(ctx)`, `Write(ctx, text)` |
| `translation` | Current translation, clipboard | `Current()`, `Retranslate()`, `Copy()` | `TranslateSelection()`, `Show(entry)`, `OnTranslated(hook)` |
| `history` | Limit, cursor | `Navigation()`, `Previous()`, `Next()` | `Latest()` |
| `store` | SQLite connection, repositories | – | `NewHistory()`, `NewLanguage()` |
| `system.IntegrationService` | unchanged | – | – |

- **Registration:** a type is registered with Wails only if the frontend calls it or it needs
  `ServiceStartup`/`ServiceShutdown`. Every exported method the frontend must not reach carries `//wails:ignore`.
  `store` and `config` are registered for their lifecycle only; `config` never exposes provider
  configuration. `system.FileService` is not registered.
- **No window service.** The frontend hides itself with `Window.Hide()` from `@wailsio/runtime` (verified).
  Showing stays in Go.
- `core`, `transport`, the DTOs and `history.History` are removed.

### Dependencies

```
models · system                                   (bottom, unchanged)
store        → system, models
config       → system
provider     → config
language     → store, config, provider, system  ← provider reports switches through a hook; `Fetcher(ID)` takes no lock
clipboard    → –                                ← no lock around reads and writes: the programs are separate
                                                  processes, and the library serializes its own calls
translation  → provider, language, clipboard   ← history records through a hook
history      → store, config, provider, language, translation
```

- Calls follow the arrows. A hook is a function registered at construction, so the package that reports
  never imports the one it reports to.
- A hook receives everything it needs as arguments and never calls back. Locks are therefore only taken in
  graph order.
- Interfaces are declared by the consumer, listing only the methods it needs.

`main.go` only wires:

```go
show := func() { window.Show(); window.Center() } // the *WebviewWindow registerWindows created

actions := system.Actions{
    Translate: func() { show(); translations.TranslateSelection() },
    Open:      func() { show(); history.Latest() },
}
```

## Startup, errors and logging - Half done

- **Registration order:** notifications, store, config, provider, language, clipboard, translation, history,
  integration.
  Wails starts services in this order and shuts them down in reverse.
- **Logger:** *(done)* `system.NewLogger()` configures the `log` package once per operating system (Linux:
  stderr, no timestamp, because the journal adds its own) and returns `slog.Default()`, so every call answers the
  same logger. `main.go` passes it as `application.Options.Logger`, before `application.New`.
  - Reason: Wails' production logger discards everything (`logger_prod.go`).
  - Inside the Wails lifecycle, everything logs through `app.Logger`. A service that does not hold the app is
    handed `app.Logger` by `main.go`.
  - `system.NewFileService()` takes no logger; it calls `NewLogger()` itself, because the CLI builds one before
    any app exists. The CLI's own output stays on `fmt`.
  - Configured before `application.New` and not in a service: Wails logs `Platform Info` from `application.New`
    and some errors from `Run` before any `ServiceStartup` (verified in a production build).
  - Scope: new and moved code logs through it. Untouched packages keep `fmt.Printf` for now.
- **`ServiceStartup` never returns an error**, except in the store, `config` and `clipboard` (no program to reach
  the clipboard is installed): a non-nil return shuts the app down.
  - It runs on the main OS thread before the GTK loop, so it must not wait for anything that needs the loop.
- **Fatal failure** (store cannot open or migrate, config file broken or just written): send a desktop
  notification through Wails' `services/notifications` and return the error from `ServiceStartup`, without
  logging it. Wails then shuts down every service already started, and `main.go` logs the error from `app.Run()`
  once and exits non-zero.
  - The notifier connects to D-Bus only in its own startup, which is why it is registered first.
  - A failure inside `application.New` can only reach the log.
- **"Not configured yet" is not fatal** once the file exists. `provider` answers with the configuration error until
  a default provider can be built. `language` and `translation` fail through it with the same message.
- **Never call a service inside `InvokeSync`/`InvokeAsync`.**
  - Verified: no other Wails callback (bound methods, shortcut, tray, menu, second instance, event listeners,
    window hooks) runs on the UI thread.
  - Services emit events while holding their lock. Once 64 events are queued, an emitting goroutine waits for
    the UI thread, so a UI-thread caller waiting for that lock would deadlock.

## Store - Done

- **File:** `~/.local/share/quick-translate/quick-translate.db` (`system.Database`), always opened, whether or not
  the history is enabled.
- **Connection:** opened in the store's `ServiceStartup` (WAL, `busy_timeout`, `foreign_keys`,
  `SetMaxOpenConns(1)`) and closed in its `ServiceShutdown`. A failed migration closes it again before the error is
  returned.
- **`store.New(files, logger)`** takes the logger, which logs with `category=store`; each repository adds
  `repository=<name>` and the migration `step=migration`.
- **Repositories** are concrete types that return `models`; no generated type leaves `store`. `NewHistory()`
  and `NewLanguage()` carry `//wails:ignore` and return nil before `ServiceStartup`.
  - `History`: `Append(*t)` sets ID and creation time, `Previous(ID, window)`, `Next(ID)`, `Latest()`,
    `HasPrevious(ID, window)`, `HasNext(ID)`. A translation that is not stored yet (ID 0) has no previous one;
    the caller steps back through `Latest()` instead.
  - `Language`: `Get(ID)` returns the languages, empty when none are stored. `Put(ID, langs)` replaces the
    provider's rows, records now as the fetch time and deletes the tags neither a provider nor the history
    references, in one transaction. `FetchedAt(IDs)` returns the fetch time of each given provider; one missing
    from the map was never fetched. `Delete(ID)` deletes the provider's languages and the orphaned tags, but
    keeps the `providers` row and its fetch time.
  - `Provider`: removed together with `models.Provider`, which only carried the fetch time. Part 2 brings the
    repository back for the settings and for deleting a provider's row.
- **Errors:** a lookup without a row returns `store.ErrNotFound` in the chain (`errors.Is`), never
  `sql.ErrNoRows`, so no caller needs `database/sql`. Errors are lowercase and wrapped once per layer.
- **Transactions** only where several statements must apply together (`Put`, `Delete`, every migration).
  Single reads and writes run directly. Nothing inside a transaction may use the `*sql.DB` or the plain queries:
  with one connection that deadlocks.
- **Language cache**, normalized; a new field on `models.Language` means a new migration:
  - `providers (id, languages_updated_at)`: the id is the `models.ProviderID` enum, never `iota`, because the
    history references it. The fetch time lives here because the provider id dictates its lifetime; the settings
    join this row in Part 2 and are updated independently of it.
  - `languages (tag)`: every tag any provider or translation uses, shared between providers.
  - `provider_languages (provider_id, tag, name, source, target, stable)`, deleted with its provider.
  - Timestamps are Unix seconds (`INTEGER`), turned into `time.Time` in the repositories.

  `language` keeps the TTL (7 days) and the background refresh. The TTL marks a list as stale, never as deleted:
  - Deleting by age would save a few rows, but force a blocking fetch when a provider is used again after a long
    time, and leave nothing to show when that start is offline.
  - *(done)* `language.refresher` runs in one goroutine, started in `language.ServiceStartup`. It makes a pass at
    start and one per provider switch, so a process that is never restarted still refreshes. Triggers coalesce.
  - A pass asks `FetchedAt` for the configured providers, fetches the stale and the never-fetched ones, at most 3
    at a time, and stores each through `Put`. A failed or empty fetch keeps the old list.
  - A refreshed list of the current provider replaces the running catalog, keeping the selection where it can.
  - Languages are deleted only with their provider (Part 2). A provider removed from `config.yml` keeps its rows
    until then; no reconcile at startup is planned.
  - The refresher runs on the application context, which Wails cancels before any `ServiceShutdown`;
    `language.ServiceShutdown` waits for it to end before the store closes.
- **History:** `translations` references the provider and both tags. `source_tag` is the detected language when
  `detection` is set and NULL if it could not be resolved; without detection it is required.
- **SQL with `sqlc`:**
  - Queries live in `internal/store/queries/<repository>.sql`; the generated code goes to `internal/store/db/`.
  - The generated code is tracked, CI checks that it is up to date, `sqlc` 1.31.1 is pinned in
    `.github/actions/setup-sqlc`, and the task `common:generate:db` regenerates it.
  - The three provider id columns are overridden to `models.ProviderID`.
  - sqlc drops `PRAGMA` statements, so `user_version` is read and written by hand in the migration.

### Migrations

```
internal/store/migrations/
  0001_initial.sql
```

- One SQL file per migration, embedded with `//go:embed`. The number prefix is the version. The same files
  are the schema `sqlc` reads.
- On every start, `migrate()` reads `PRAGMA user_version` and runs every higher-numbered file in ascending
  order.
- **Each file runs in one transaction together with `PRAGMA user_version = N`.** SQLite rolls `user_version`
  back with the transaction (verified), so a failed migration is retried on the next start.
- **Checked before running:** only `.sql` files named `<number>_<name>.sql`, starting at 1 with no gaps and no
  duplicates. Otherwise every start fails.
- **A `user_version` above the newest file is refused**, with the hint to update or run `--purge`.
- **Up only, SQL only.** Go migrations get added to the engine only when a data migration needs one.
- **No migration from before the store:** `0001` creates the schema from scratch.

## Models - Open

Postponed until the other packages are moved onto the services. `internal/models` holds `ProviderID` with its slug
mapping, `Language` and `Translation`; the stored `Provider` is removed *(done)*. The old `Provider` interface was
replaced inside `provider` *(done)*; `core` does not build until it is removed.
`LanguagePreferences` moved to `config` *(done)*.

- **Constraint:** `ProviderID` must stay in a package `store/db` can import without a cycle; `store` itself is
  not one (`store → db → store`).
- **Constraint:** `Language` must not come from `store/db` while `provider` produces it and `db` imports the
  package `ProviderID` lives in.
- **Provider interface** *(done)*: providers implement the unexported `provider.implementation`
  (`translate(ctx, Request) (Result, error)`, `languages(ctx)`) and get one row in `registry`. Consumers get two
  immutable wrappers: `*Translator` (the current provider, swapped on `Select`) and `*Fetcher` (built on demand
  for any configured provider). Neither takes a lock, so requests never run under the provider lock and several
  fetchers can run at once. The package hands every provider the shared HTTP client with the request timeout.
  Settings are decoded through the config decoder and validated when a provider is built.

## Files (`internal/system`) - Done except the notification file

- **`system.AppName = "Quick Translate"`**, next to `ApplicationID`. *(done)*
  - `main.go` takes `Options.Name` and the window title from it, and `TrayLabel` becomes `AppName`.
  - The constant is what keeps `FileService` and the CLI free of the app instance.
- **Rows:** *(done)*
  - `History` becomes `Database`.
  - `DataRoot` is reported as **"Application data"**.
  - `CacheRoot` and `Languages` are removed, with no transition.
  - New: a directory row **"Notifications"** at `filepath.Join(xdg.ConfigHome, AppName)` (category `data`), with
    `notification-categories.json` inside it as an `internal` row.
- **Why the Notifications rows exist:** the Wails notifier writes that file on every shutdown, into
  `os.UserConfigDir()/<Options.Name>`, a path that cannot be configured. It creates the directory `0755` and
  the file `0644`.
  - Fix: `system` creates the directory with `0700` and the file with `0600` (content `{}`) before the service
    starts, and Wails keeps those modes. *(open: lands with the notifications service in step 5)*
- **`Remove` checks rows instead of names.** *(done)* A directory is removed only if its path is the path of a
  row marked `directory`.

## Configuration (`config`) - Done

| Consumer | Reads |
|---|---|
| `provider` | `DefaultProvider`, `Providers`, `Provider(ID)` |
| `language` | `Preferences` |
| `history` | `HistoryLimit` |

Provider configuration goes through a decoder, because a Go method cannot be generic:

```go
func (s *Service) Provider(ID models.ProviderID) (decode func(into any) error, err error)

decode, err := s.config.Provider(models.DeepL)
instance := provider.NewDeepl()
err = decode(instance)
```

- The error from `Provider` means not configured; the error from `decode` means configured but unreadable.
- Today the decoder is `yaml.Node.Decode`. In Part 2 the settings repository backs the same signature.
- The default provider is read once at startup. `Select()` changes only the running process.
- **Startup:** `ServiceStartup` writes the template if the file is missing and then fails, so the user fills it
  in before the app starts. Otherwise it reads, decodes and validates the whole file once (a negative
  `max_entries` is broken) and keeps it in memory; edits need a restart. A broken file fails startup.
- `Preferences()` and `HistoryLimit()` return no error, since the file is validated at startup. `Providers()`
  returns the configured providers sorted by ID.
- **Provider IDs, not slugs:** the file keeps the slugs; `load()` resolves them to `models.ProviderID` through
  `models.ProviderIDFromSlug`, and an unknown slug is a broken file. `ProviderID.String()` returns the slug.
- **Errors** are lowercase and prefixed with `config: `, like the store's.
- Every getter carries `//wails:ignore`.

## Provider switch

`provider.Select(ID)` runs as one sequence while holding the provider lock:

1. **Build the provider** through the config decoder and list the configured providers for the event. If either
   fails, stop.
2. **Build the language list:** call the hooks, in the order they were added, with a `*Fetcher`, whose `Info()`
   carries the ID, the name and whether it detects. The first error stops the switch. `language` reads the stored
   list and fetches over the network only if none is stored, so a switch to a known provider works offline. If it
   fails, `language` keeps its state and emits nothing, and the hook returns the error.
3. **Commit and emit only if both succeeded:**
   - `language` swaps its `Collection` and emits `languages:changed`, then `selection:changed`.
   - After the hook returns, `provider` swaps its translator and emits `provider:changed` *(done)*.

`Current()` and `Translator()` wait while a switch is running; that is accepted in exchange for consistency.
`Fetcher(ID)` never waits: the background refresh builds one per stale provider from `FetchedAt` and may run them
concurrently.

A translation reads `provider.Translator()` and `language.Snapshot()` separately. The snapshot carries the ID
its `Collection` was built for, the translator carries it in `Info()`.
- If the two IDs differ, it reads both again. The second read waits for the switch to finish.
- *(done)* The `Collection` and its builder are replaced by an immutable catalog per provider, which carries the
  ID and builds its matchers once, and a `Selection` value. Detection is a boolean (`Selection.Detection`), no
  longer the virtual `"auto"` tag. Every rule maps a selection to a new one, so a failure changes nothing.
- The locale the target language falls back to comes from `system.IntegrationService.Locale()`, read once.

## Translation and history

- `translation` reports every new translation through `OnTranslated`. `history` appends it, moves the cursor
  and emits `history:changed`.
- `history` shows a stored entry by calling `translation.Show(entry)`, which does not fire the hook.
- The cursor (today `core.translation.ID`) moves into `history`.
- **A history step** calls provider, language, translation and history in that order, and each emits its own
  events. `history.Latest()` works the same way, which replaces the `restored` event.
- *(done)* `translation` holds two locks. `work` serializes translations from `translation:started` to
  `:changed`/`:failed`; `mutex` guards the current translation and the hooks and is never held during a request,
  so `Current()` and `Show()` never wait for the network. Lock order: `work` → provider, language, the hooks →
  `mutex`. The hooks run after `translation:changed`, holding `work` but not `mutex`, so `history` may call
  `Show()` under its own lock.
- *(done)* An unchanged repeat (same text, provider, detection, source and target) answers with the current
  translation: no request, no hook, but still `translation:started` and `:changed`.
- *(done)* While detecting, the stored `SourceTag` is the detected language as `language` resolved it, read back
  through `Snapshot()` after `SetDetected`; nil if it is unknown.

## Events

State reaches the frontend **only** through events. Mutations return nothing but an error; getters are only
for the first load.

| Event | Payload | Emitted by |
|---|---|---|
| `languages:changed` | `language.Options{Detection, PreferredSource, PreferredTarget, SourceLanguages, TargetLanguages}` | language, on a rebuild |
| `selection:changed` | `language.Selection{Detection, Source, Target, Detected}` | language |
| `provider:changed` | `provider.State{Current, Providers}` | provider |
| `translation:started` / `:changed` / `:failed` | –, the translation, the message | translation |
| `history:changed` | `history.Navigation{HasPrevious, HasNext}` | history |

- **Emission order is part of the contract** (the table's order). Whatever a later event is resolved against
  comes first, so a selected tag always finds its name in the list.
  - Verified: Wails delivers every event, in emission order, whatever its size.
- **Emit while holding the lock of the state**, so that emission order equals mutation order.
- **Every `translation:started` gets exactly one `:changed` or `:failed`**, even when `Retranslate()` also
  rejects its promise.
- **Per service on both sides:**
  - Go: name constants in `service.go`, registered with `application.RegisterEvent` in `init()`. It is not done
    in the constructor, because `RegisterEvent` panics on a duplicate name.
  - Frontend: the names live in the composable, which has a `listen()` that returns its own unsubscribe.
- **A typed `on()` helper in `lib/`** accepts only `keyof Events.CustomEvents`. `Events.On` accepts any string
  and silently turns `data` into `any`.

## Frontend

- **Never read state right after `await`.** A promise resolves before the events of its own call arrive
  (verified; always for payloads over 8 KB, such as `languages:changed` at 17.6 KB). Components are reactive
  and unaffected.
- **Pending comes from `translation:started` → `:changed`/`:failed`**, for every source of a translation. The
  request counter and the shortcut flag in `request.ts` go.
- **The frontend composes use cases and retranslates explicitly;** no service translates as a side effect:

  ```ts
  await request(async () => { await SetSource(tag); await Retranslate(); });
  ```

- `services/events.ts` and `services/wire.ts` go. Each composable maps its own events onto its own state group.
- `useApplication.start()` calls every domain's `listen()`, then the getters.
- Routing reacts to `translation:started` and `translation:changed`.
- `useApplication.hide()` uses `Window.Hide()`.

## Upgrading from 0.1.0-alpha

Users run `quick-translate --purge` **with the installed version before installing the new one**, then write
the configuration again. The old binary's purge still knows `history.db`, the gob cache and `config.yml`,
including the provider key. This goes into the release notes and `build/linux/INSTALL.md`.

## Order of work

1. **Logger and `system`:** *(done)* `AppName`, the file rows, `Remove` by row, `system.NewLogger()`. Only the
   notification file is left for step 5.
2. **`store`:** *(done)* migrations, `sqlc`, the history, language and provider repositories. `history.History`
   and the gob cache go once their consumers use the repositories (step 4).
3. **`config.Service`:** *(done)* the YAML file read once in `ServiceStartup`, getters on the service, provider IDs
   instead of slugs. `core` keeps calling the removed free functions until step 4 replaces it.
4. **Backend services with the adapter as a facade.** The adapter delegates and keeps building the DTOs, so the
   frontend stays untouched while locks and hooks are checked.
5. **Events and getters, notifications service, registration order.** The adapter and the new bindings run side
   by side.
6. **Frontend, domain by domain:** history, translation, languages and provider, first load.
7. **Delete** the adapter, DTOs, `core` and `transport`; regenerate the bindings.
8. **Docs:**
   - `architecture.mmd`
   - `CLAUDE.md`: architecture, runtime files, commands, conventions
   - `README.md`, `SECURITY.md` (the database always exists and holds the cache), `CONTRIBUTING.md`
   - `build/linux/INSTALL.md`

**Out of scope:** pruning the history. It is a log that is never cut; the limit only bounds how far back you can
step.

# Part 2 — Dissolving the `config` package

## Why

Today the user must write a file by hand before anything works. It holds the provider key in plain text and
is only read at startup, so every change needs a restart. Settings move into the store, get validated as they are entered, and get migrations
for free.

## Settings in the store

- **Store:** one more repository in `store`, and `config.Service` reads through it instead of the YAML file.
- **Same interface:** the decoder signature from Part 1 stays, so consumers don't change.
- **Validation:** settings are validated as they are written.

## Secrets in the OS keyring

- **Keyring:** freedesktop Secret Service over D-Bus, through `github.com/zalando/go-keyring` (no cgo). `godbus`
  is already a dependency.
- **Behind an interface**, like `clip` and `models.Provider`. CI has neither D-Bus nor a wallet.
- **Read lazily on the first translation, with a retry, and cache the key in memory.**
  - The autostart unit has no ordering against the secret service, so the keyring may not be up yet at login.
  - An unlock dialog should appear right after the shortcut, not unprompted at login.
- **No keyring** (bare sway, i3, Hyprland): fall back to a plain `0600` file.
  - The settings view says where the key lives ("OS keyring" versus "local file, unencrypted").
  - The fallback is hand-written, about 40 lines.

## Invalidation on change

Each consequence is a hook that `config` calls on the owning service:

| Change | Consequence |
|---|---|
| Provider auth key | Invalidate that provider's language cache |
| Provider removed | `LanguageRepository.Delete`, then the provider's row through `ProviderRepository` |
| Default provider | `provider.Select` |
| `history.max_entries` 0 → >0 | `history` starts recording |
| Language preferences | `language` resolves them again; the frontend retranslates |

## Entry points

A settings view is opened through `Exec=quick-translate --settings`, a `[Desktop Action Settings]` in the desktop
entry, and is recognized in `OnSecondInstanceLaunch` through `SecondInstanceData.Args`.

## Escape hatch

Without the text file, there is no way to configure the app from dotfiles, and no way in when the database
breaks. Add:

- `--reset-settings`
- Export and import of the settings as YAML or JSON over the CLI
