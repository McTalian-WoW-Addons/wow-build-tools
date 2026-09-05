# wow-build-tools decisions

Rationale moved out of `CLAUDE.md` when it was rewritten to the `wow-dev` plugin's
≤2048-byte contract (`## Commands` / `## Conventions` / `## Docs` only). Not read by
any skill or agent — background for humans.

## Why `make test` and `./trunk check` before every commit, not just the git hook

`trunk git-hooks sync` installs a local pre-commit hook (points `core.hooksPath` at
`~/.cache/trunk/...`, scoped to this repo only — a fresh clone needs it re-run), but
it only runs `trunk fmt` (autofixers). Non-autofixable issues — e.g. a markdownlint
content rule like `MD041` — pass silently through commit and only surface at
`trunk-check-pre-push` or CI. Nothing runs the Go test suite on commit or push at
all, so `make test` has to be run by hand (or via `/wow-dev:run-checks`) before
committing.

## Why per-package Renovate abandonment suppression, not a global threshold edit

Renovate flags abandoned/dormant dependencies on the [Dependency Dashboard
issue](https://github.com/McTalian-WoW-Addons/wow-build-tools/issues).
`renovate.json`'s `abandonmentThreshold` (inherited from `config:best-practices`)
drives that. A single actively-maintained-but-slow-releasing dependency triggering a
false positive should not lower the org-wide bar for catching genuinely abandoned
ones — so a false positive is suppressed per-package with a `packageRules` entry
setting `"abandonmentThreshold": null`, leaving the global threshold intact.

## Why the dependency-swap-prober agent exists

Before merging a dependency swap (fork migration, replacement library, etc.),
`go test`/CI only re-runs assertions that already exist, and won't catch drift in
code paths nothing currently asserts on — custom `(Un)MarshalYAML`, CLI help text,
or anything hand-rolled around the dependency's API. The
**dependency-swap-prober** agent (`.claude/agents/dependency-swap-prober.md`)
exists to validate behavioral parity that the test suite structurally can't.
