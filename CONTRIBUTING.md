# Contributing

Thank you for looking at Quick Translate. The project is in **alpha** and maintained by one person, so for
anything larger than a small fix, please ask first:

- **[Open an issue](https://github.com/manuelschoene/quick-translate/issues)** (preferred, the answer helps
  the next person too)
- **Email:** [schoene-manuel@gmx.de](mailto:schoene-manuel@gmx.de)

## Getting set up

The requirements are listed in the [README](README.md#from-source). Then:

```sh
task dev                        # run with live reload
task build                      # -> bin/quick-translate
task build DEV=true             # faster: no production tags, no compression
task archive                    # -> bin/*.tar.gz, everything a release ships
task art                        # re-render art/tray.png and art/wordmark.svg (needs Inkscape)
task common:generate:bindings   # regenerate frontend/bindings/
task --list                     # everything else
```

The frontend uses **bun**, not npm. From `frontend/`:

```sh
bun run types    # vue-tsc: the type check
bun run lint     # ESLint, fixes what it can
bun run format   # Prettier
```

## There are no tests

The repository has **no test suite**, so nothing catches a regression for you. **Verify your change by running
the application.** Tests are very welcome as a contribution of their own: `internal/language/matching.go`,
`internal/system/kde.go` and `internal/history/db.go` would profit first.

Take `bun run types` seriously. `verbatimModuleSyntax` is on, so a type imported as a value breaks the whole
frontend at runtime, and the type check is the only thing that notices. `bun run build` does not run it.

## Code conventions

Matching the code around your change matters more than any rule below.

- **Go:** a prose doc comment above every function, unexported ones included, that describes behavior and
  edge cases rather than the signature.
- **Error strings are GUI text:** capitalized full sentences that tell the user what to do. Wrap with `%w`
  when the cause matters.
- **Failures that only cost a feature** are printed and swallowed. Only failures that make the application
  unusable are returned.
- **Frontend:** the same prose style in TSDoc. Components only talk to composables, and `services/wire.ts` is
  the only module that knows Go field names.
- **Files and paths:** every path and permission comes from the table in `internal/system/files.go`. Never
  build a path of your own.
- **Platform-specific code** goes behind build tags, like `internal/clipboard/linux.go`.
- **Import sorting is ESLint's job.** Do not add an import-sorting Prettier plugin.
- **Commit generated files:**
  - `frontend/bindings/`, after changing an exported method on `transport.Adapter`, a DTO or a registered
    event. CI fails when they are stale.
  - `art/tray.png` and `art/wordmark.svg`, after changing a source in `art/`.
- **`architecture.mmd`:** update it when backend types or package boundaries change.

## Using AI assistants

Welcome, and nothing to disclose. But there are no tests, so before opening a pull request:

- **You can explain every line.**
- **You have run it:** built, installed, used the feature.
- **It follows the conventions above.** Assistants tend to drift from the doc comments and the error strings.

`CLAUDE.md` carries the architecture and conventions in a form assistants read well.

## Pull requests

Branch off `main` and open a pull request back into it. CI only runs on branches named `feature/…`,
`hotfix/…`, `chore/…`, `docs/…`, `refactor/…` or `test/…`. `main` is protected and must stay installable.

**The title must be a [Conventional Commit](https://www.conventionalcommits.org/).** release-please reads it
to write the changelog and decide when a release is due:

```
feat: Added keyboard navigation in language selection
fix: Fixed inconsistencies with ESLint
feat!: Changed the configuration format     # ! marks a breaking change
```

Common types: `feat`, `fix`, `refactor`, `docs`, `chore`, `perf`, `test`, `build`, `ci`.

In the description, say **how you verified the change**. For the desktop integration, the clipboard or the
installation, name your distribution, desktop and session type (Wayland or X11). The pull request template
scaffolds this.

### What CI does

1. **Fix:** runs `gofmt -s -w`, ESLint and Prettier and commits the result to your branch. It never fails.
   From a fork it cannot push, so run `gofmt -s -w .` and `bun run format` yourself.
2. **Check:** the TypeScript check, ESLint, Prettier, the frontend build, gofmt, `go vet`, `go test`,
   `go mod verify`, an unchanged `go mod tidy` and `govulncheck`. Then it builds the release archive
   (`task archive UPX=false`), validates the desktop entry, checks that `frontend/bindings/` is current and runs
   `--version`, `--help` and `--status` on the built binary.

## Reporting bugs

[Open an issue](https://github.com/manuelschoene/quick-translate/issues/new/choose) with the bug report form.
The desktop integration and the clipboard behave differently on nearly every setup, so the form asks for
`--version`, `--status`, your distribution, desktop and session type. For security problems, read
[SECURITY.md](SECURITY.md) first.

## License

By contributing you agree that your contribution is licensed under the [Apache License 2.0](LICENSE).
