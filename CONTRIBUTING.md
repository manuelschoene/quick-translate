# Contributing

Thank you for looking at Quick Translate. Contributions are welcome.

The project is in **alpha** and maintained by one person, so please read the two sections below before you
start writing code — they will save you the most time.

## Questions go to me

There is no team and no triage rota. Anything you are unsure about — whether a feature fits, whether a bug
is already known, whether an approach is the one I would take — please ask me directly:

- **[Open an issue](https://github.com/manuelschoene/quick-translate/issues)** — preferred, because the
  answer is then useful to the next person as well.
- **Email:** [schoene-manuel@gmx.de](mailto:schoene-manuel@gmx.de)

For anything larger than a small fix, asking first is genuinely worth it. The project has a direction that
is not fully visible from the code yet.

## There are no tests

This is the part to be aware of. **The repository contains no test suite** — no `*_test.go`, no frontend
test runner, nothing you can run to find out whether your change broke something. That is a known gap and
it is on the list for the stable release, but today it means:

- **Nothing catches a regression for you.** Please verify your change by running the application.
- If you want to add tests, that contribution is very welcome on its own. `internal/language/matching.go`,
  the KWin rule merge in `internal/desktop/kde.go` and the window queries in `internal/history/db.go` are
  the places where tests would earn their keep first.

What *is* available are the checks that find errors without running anything. Continuous integration runs
all of them on your pull request, but running them yourself first is far quicker than waiting for a red
build.

### Backend

```sh
go vet ./...     # the only automated backend check there is
gofmt -l .       # must print nothing
make build       # compiles the frontend and the application
```

### Frontend

Run these from `frontend/`. The package manager is **bun**, not npm.

```sh
bun install
bun run types    # vue-tsc --noEmit — the type check, and the closest thing to a test suite
bun run lint     # ESLint: import order, unused imports, and it fixes what it can
bun run format   # Prettier: 4-space indent, single quotes, 120 columns, Tailwind class sorting
```

`bun run types` is the one to take seriously. `verbatimModuleSyntax` is on, so a type imported as a value
resolves to a missing export in the browser and takes the whole module graph down at runtime — the type
check is what catches that, and nothing else will.

Note that `format` and `lint` do different jobs and neither replaces the other: import sorting is an ESLint
concern here, not a Prettier one. Please do not add an import-sorting Prettier plugin — the two would fight.

## Getting set up

The prerequisites, the build and the installation are described in the [README](README.md). In short:

```sh
make dev         # run with live reload
make build       # -> build/bin/quick-translate; UPX=0 skips the slow compression
make install     # build, install the binary, register it with the desktop
make help        # every target and variable
```

The WebKit build tag is detected with `pkg-config`, so you never write it by hand.

## Code conventions

The codebase is fairly consistent; matching what is around your change matters more than any rule below.

- **Go:** a full-sentence prose doc comment above every function, including unexported ones. Describe what
  it does and which edge cases it handles, rather than restating the signature.
- **Error strings are user-facing GUI text.** Capitalised, punctuated, and where possible they tell the
  user what to do: `"No target language is set. Please choose the language you want to translate into."`
  Wrap with `%w` when the cause matters.
- **Failures that only cost a feature** (an unreadable history, an unresolvable detected language) are
  printed and swallowed. Only failures that make the application unusable are returned as errors.
- **Frontend:** the same prose style in TSDoc blocks. Components and views only ever talk to composables;
  `services/wire.ts` is the only module that knows Go field names.
- **Platform-specific code** goes behind build tags, in the same shape as `internal/clipboard/linux.go`,
  `internal/transport/unix.go` and `internal/desktop/linux.go`.
- **`architecture.mmd`** is a Mermaid class diagram of the Go packages. Please update it when backend types
  or package boundaries change.
- After changing an **exported** method on `transport.Adapter` or a DTO, regenerate the frontend bindings
  in `frontend/wailsjs/` and **commit them with your change**. They are generated but tracked, so that a
  checkout type-checks and lints without a Go toolchain — and CI fails if they are out of date.

Regenerating those bindings has one trap worth knowing about. Only `wails build` and `wails dev` write
them — `wails generate module` runs, prints nothing and writes no files — and neither can do it while Quick
Translate is running. The generation step builds the application and executes it, and that process hands
the request over to the instance already holding the socket and exits before the bindings are written. The
build then reports `Generating bindings: Done.` without having written anything and fails on the missing
modules. Stop it first:

```sh
systemctl --user stop quick-translate.service
make build
systemctl --user start quick-translate.service
```

## Using AI assistants

AI assistants are welcome here, and there is nothing to disclose or flag when you use one. A good patch is
a good patch regardless of how it was written.

What does not work is code that nobody has read. Before you open a pull request, please make sure that:

- **You can explain every line** — what it does and why it is there.
- **You have actually run it.** Built it, installed it, used the feature. Not "the diff looks reasonable".
- **It follows the conventions above.** Assistants drift away from them, particularly the prose doc comment
  on every function and the error strings that are written as user-facing GUI text.

The reason is the section further up: there are no tests. Nothing in this repository catches a change that
looks plausible and is subtly wrong, so the only thing between that and a release is a person having
understood the code. If you cannot answer a question about your own pull request, it is not ready yet.

`CLAUDE.md` in the repository root carries the architecture and these conventions in the form assistants
read best — pointing yours at it will save you most of the drift.

## Pull requests

Branch off `main`, name the branch `feature/…` for anything new or `hotfix/…` for an urgent fix, and open
a pull request back into `main`. There is no long-lived development branch: `main` is protected, every
change arrives through a pull request, and `main` is what people clone and build — so it has to stay
installable at all times. Released versions are downloaded from the
[releases page](https://github.com/manuelschoene/quick-translate/releases); `main` is the nightly state.

Continuous integration runs on every push to a `feature/…` or `hotfix/…` branch and on every pull request:

- The first job **fixes** rather than checks. It runs `gofmt -s -w`, ESLint and Prettier and commits the
  result back to your branch, so a forgotten formatting run is not something you have to fix by hand. It
  never fails the build.
- The second job **checks**: `gofmt`, `go vet`, `go test`, `go mod verify`, a `go mod tidy` that has to
  leave `go.mod` and `go.sum` unchanged, `govulncheck`, the TypeScript check, ESLint, Prettier, a full
  `make build`, and `desktop-file-validate` over the desktop entry the binary generates.

One thing to know if you work **from a fork**: the auto-fix job cannot push to your branch, because a pull
request from a fork gets a read-only token. It is skipped rather than failing, so please run
`gofmt -s -w .` and `bun run format` yourself before opening the pull request.

**Use [Conventional Commits](https://www.conventionalcommits.org/) in the pull request title.** This is the
one hard requirement, because the titles are what release notes will be generated from:

```
<type>: <description>
```

Common types: `feat`, `fix`, `refactor`, `docs`, `chore`, `perf`, `test`, `build`, `ci`.

Examples from this repository:

```
feat: Added keyboard navigation in language selection
fix: Fixed inconsistencies with ESLint
refactor: Sharpened backend interface and restructured frontend logic
```

Add a `!` after the type for a breaking change (`feat!: …`) and say in the description what breaks.

In the body, please describe **how you verified the change** — since nothing is automated, that is the only
signal I have. "Ran `make install`, pressed the shortcut on Wayland/KDE, translated a selection, stepped
through the history" is worth far more than "works".

If your change touches the desktop integration, the clipboard or the installation, please say which
distribution, desktop environment and session type (Wayland or X11) you tested on. Those paths differ a lot
between setups and I can only test my own.

## Reporting bugs

[Open an issue](https://github.com/manuelschoene/quick-translate/issues/new/choose) and pick the bug
report. The form asks for what makes a report actionable, and it is worth having ready:

- The output of `quick-translate --status`, which names the build and shows what is installed where
- Your distribution, desktop environment and session type (Wayland or X11)
- Which clipboard tool you have installed (`wl-clipboard`, `xclip` or `xsel`)
- The journal, `journalctl --user -u quick-translate.service -n 50`, which also reports which clipboard
  backend the application picked

None of that is bureaucracy: the desktop integration and the clipboard behave differently on nearly every
setup, and without those details a report usually cannot be reproduced at all.

For anything that looks like a security problem, please read [SECURITY.md](SECURITY.md) first.

## License

By contributing you agree that your contribution is licensed under the
[Apache License 2.0](LICENSE), the same as the rest of the project.
