# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

The caveman response-style directive lives in `AGENTS.md` at repo root (also loaded automatically) — applies here same as the rest of the workspace.

## What this repo is

Go CLI (`github.com/McTalian/wow-build-tools`, Cobra commands in `cmd/`, implementation in `internal/`) that packages and uploads WoW addons to CurseForge/WoWInterface/Wago, and ships as a reusable GitHub Action (`action.yml`) other addon repos in this workspace depend on for CI. Also hosts the shared i18n Python scripts (`scripts/i18n/`) those repos clone at build time. See `~/code/CLAUDE.md` for how this fits into the wider workspace.

## Commands

```bash
make build    # dist/ binaries for linux + windows, into .local-dist/
make test     # scripts/test.sh — go test plus a cyclomatic-complexity gate (CC_THRESHOLD=10)
make tools    # install Go tools from go.tools; requires jq
make release  # copy built binaries onto PATH (~/bin and the Windows-side bin)
```

## Dependency swaps

Renovate flags abandoned/dormant dependencies on the [Dependency Dashboard issue](https://github.com/McTalian-WoW-Addons/wow-build-tools/issues). `renovate.json`'s `abandonmentThreshold` (inherited from `config:best-practices`) drives that; suppress a false positive per-package with a `packageRules` entry setting `"abandonmentThreshold": null` rather than editing the threshold globally.

Before merging a swap (fork migration, replacement library, etc.), use the **dependency-swap-prober** agent (`.claude/agents/dependency-swap-prober.md`) to validate behavioral parity — `go test`/CI only re-runs assertions that already exist, and won't catch drift in code paths nothing currently asserts on (custom `(Un)MarshalYAML`, CLI help text, anything hand-rolled around the dependency's API).
