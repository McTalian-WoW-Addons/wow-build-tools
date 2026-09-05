# wow-build-tools

Go CLI (`github.com/McTalian/wow-build-tools`) that packages and uploads WoW addons to CurseForge/WoWInterface/Wago, and ships as a reusable GitHub Action other addon repos in this workspace depend on for CI. Also hosts the shared i18n Python scripts (`scripts/i18n/`) those repos clone at build time.

## Commands

- `make build` — dist/ binaries (linux + windows) into `.local-dist/`
- `make test` — `scripts/test.sh`: go test + cyclomatic-complexity gate (`CC_THRESHOLD=10`)
- `make tools` — install Go tools from `go.tools`; requires `jq`
- `make release` — copy built binaries onto `PATH` (`~/bin` + Windows-side bin)

## Conventions

- Run `make test` and `./trunk check` before every commit; the local pre-commit hook only autofixes (`trunk fmt`), it doesn't catch everything.
- Suppress a false-positive Renovate abandonment flag per-package (`packageRules` + `"abandonmentThreshold": null`), never edit the global threshold.
- Validate dependency swaps (fork migration, replacement library) with the **dependency-swap-prober** agent (`.claude/agents/dependency-swap-prober.md`) before merging.

Full list: `docs/agent/conventions.md`.

## Docs

- `.github/docs/future-work.md` — org migration checklist and remaining cross-repo tasks.
- `docs/agent/decisions.md` — rationale behind the pre-commit-hook and dependency-swap policies above.
