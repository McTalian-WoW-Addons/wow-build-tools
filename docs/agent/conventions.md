# wow-build-tools conventions

## Structure

- Cobra command wiring lives in `cmd/` (one file per subcommand, e.g. `cmd/build.go`, `cmd/upload_curse.go`); the logic each command calls into lives under `internal/<package>` (`internal/build`, `internal/toc`, `internal/upload`, etc.) — keep `cmd/` thin.
- New shared logic gets its own `internal/<package>`; don't add a second, unrelated concern to an existing package (e.g. flavor lookups stay in `internal/flavor`, TOC parsing stays in `internal/toc`).
- Mock collaborators with a `Mock<Type>` struct holding `<Method>Func func(...)` fields (see `internal/repo/mock_repo.go`), not a hand-rolled fake or a mocking library.

## WoW API

- WoW flavor identity (id, display name, client directory) is the single table `flavor.KnownFlavors` in `internal/flavor/flavor.go`; add a new flavor there, never hardcode a flavor string or directory elsewhere.
- TOC `## Interface` parsing and interface-number-to-flavor mapping lives in `internal/toc`; confirm any interface-number band against `~/code/wow-ui-source` (ground truth) before hardcoding one.
- This tool ships as a GitHub Action (`action.yml`) other addon repos' CI depends on — a behavior change to a `cmd/` flag or an `internal/build`/`internal/upload` default is a breaking change for every consumer repo, not just this one.

## Strings

- Emit user-facing output through `internal/logger` (`logger.Info`/`Warn`/`Error`/`Success`/`Debug`, etc.), not a bare `fmt.Println`, so log level and color formatting stay consistent; a bare `fmt.Println`/`fmt.Print` is only for deliberate blank-line spacing or raw machine-readable output (e.g. `toc_update.go`'s JSON dump).
- Route CLI help/usage text through the Cobra command's `Short`/`Long`/`Example` fields, not ad hoc prints inside `RunE`.

## Testing

- Tests are plain `go test`, table-driven, using `testify`'s `assert`/`require` — see `internal/config/config_test.go`.
- Use `t.TempDir()` for filesystem fixtures; never write test scratch files into the repo tree.
- Run the suite only through `make test` (`scripts/test.sh`) — it also enforces the cyclomatic-complexity gate (`gocyclo`, `CC_THRESHOLD=10` in the `Makefile`) and generates coverage/report artifacts under `.coverage/`; a bare `go test ./...` skips both.
- `make tools` installs the pinned `gopogh`/`gocyclo`/`covreport` versions from `go.tools` that `scripts/test.sh` requires; run it once per clone before `make test`.

## Packaging

- `make build` is the only sanctioned way to produce binaries; it builds both `linux/amd64` and `windows/amd64` into `.local-dist/` so a local build never drifts the `dist/` directory that CI commits on release.
- `make release` copies the freshly built binary onto `PATH` (`~/bin` and the Windows-side bin) — re-run it after every `make build` you intend to actually use locally.
- Validate a dependency swap (fork migration, replacement library) with the **dependency-swap-prober** agent (`.claude/agents/dependency-swap-prober.md`) before merging — `go test`/CI only re-run existing assertions and won't catch drift in unassisted code paths (custom `(Un)MarshalYAML`, CLI help text, hand-rolled API glue).
