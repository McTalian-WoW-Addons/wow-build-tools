# Org Migration Session Summary

**Session Date:** February 28 – March 2, 2026 (Session 2: March 2)
**Purpose:** Migrate `Endeavoring`, `RPGLootFeed`, and `wow-build-tools` from `McTalian` personal account to `McTalian-WoW-Addons` GitHub organization.

---

## What Was Accomplished

### Phase 1: Pre-Transfer Preparation ✅

- **Org-level secrets configured:** `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`, `GH_PAT` — all set at org level, scoped to all repos.
- **Repo-level secrets identified:** `DISCO_WH_NDVRNG_RELEASES` (Endeavoring) and `DISCO_WH_RLF_RELEASES` (RPGLootFeed) stay per-repo. _(Later renamed to `DISCORD_RELEASES_WEBHOOK` in both repos — see Phase 4.)_
- **`.github` org repo created** (public) with org profile README.

### Phase 2: Repo Transfers ✅

- **New fine-grained PAT** created with `McTalian-WoW-Addons` as resource owner (old PAT still active, should be retired after full verification).
- **All 3 repos transferred** in order: `wow-build-tools` → `Endeavoring` → `RPGLootFeed`.
- **Post-transfer for each repo:**
  - Local git remote updated
  - Repo-level secrets deleted (they override org-level secrets with same name — learned the hard way)
  - Branch protection bypass lists re-added as org owners/admins (bypass lists do NOT transfer)
- **Remaining:** Verify workflows run with org-level secrets on Endeavoring and RPGLootFeed (will happen organically with next release/PR).

### Phase 3: Update Hardcoded References — Nearly Complete

Three separate branches created with reference updates. Status per repo:

#### Endeavoring

- **Branch `chore/update-org-references`:** Merged via PR #12.
- **Branch `update-more-references`:** Currently on this branch (1 commit ahead of main). Contains additional missed references:
  - `Endeavoring.toc` — X-Website URL
  - `endeavoring-1-1.rockspec` — source URL
  - `package.json` — repository URL (semantic-release uses this!)
- **PR:** Merged.

#### RPGLootFeed

- **Branch `chore/update-org-references`:** PR #527 open.
- Contains 3 commits (original update + user's additional fixes).
- **Files changed (28):** workflows, README.md, README.BB, About.lua, 13 locale files, prompts, docs, PR template, `.scripts/missing_translation_check.py`, `RPGLootFeed.toc`, `package.json`, `rpglootfeed-1-1.rockspec`.
- **Status:** PR is open, needs merge.

#### wow-build-tools

- **Branch `chore/update-org-references`:** 1 commit ahead of beta.
- **Files changed (7):** `release-published.yml`, `README.md`, `updatebin.go` (repo const), `toc/check/README.md`, `toc/update/README.md`, plus new `docs/org-migration-plan.md` and `future-improvements.md`.
- **PR:** Merged.
- **Note:** Go module path (`github.com/McTalian/wow-build-tools`) in `go.mod` and all Go imports intentionally NOT updated yet — works via GitHub redirect. Can update later as a separate change.

### Phase 3.5: Standardize Repo Rulesets ✅

Org-level rulesets require GitHub Team (paid plan, $4/user/month additive to individual Pro/Copilot). Instead, standardized repo-level rulesets across all three repos via API.

- **Standardized config applied to all three repos:**
  - Branches: `~DEFAULT_BRANCH`, `v*`, `alpha`, `beta`, `main`
  - Rules: deletion, non_fast_forward, update (block force push), creation (block new protected-pattern branches)
  - PR: 1 approval required, squash + rebase allowed
  - Status checks: "Passing PR Checks" (skip on create)
  - Copilot code review: enabled
  - Bypass: OrganizationAdmin (always) — unblocks solo maintainer workflow
- **Reusable tooling created in `.github` org repo:**
  - `rulesets/default.json` — canonical ruleset definition (single source of truth)
  - `scripts/apply-ruleset.sh` — CLI to apply to one repo or all repos (`--all`), with `--dry-run` support
  - Detects existing rulesets (updates) vs missing (creates)
- **Auth note:** `gh auth refresh -s admin:org` was needed for ruleset API access

---

## Lessons Learned / Gotchas

1. **Repo-level secrets override org-level secrets with the same name.** After transferring, the old repo-level `GH_PAT` shadowed the new org-level one. Fix: delete repo-level secrets that are now org-level.

2. **Branch protection bypass lists don't transfer.** The personal account admin bypass didn't carry over. Fix: re-add as org owners/admins.

3. **`package.json` `repository.url` matters for semantic-release.** Semantic-release uses it to determine the GitHub repo for creating releases. When you drop `package.json` per addon (Phase 5), the `repositoryUrl` in `.releaserc.json` or CI-detected repo URL will be the fallback.

4. **More reference locations than expected.** Beyond workflows, references existed in:
   - `.toc` files (`X-Website`)
   - Rockspec files (`source.url`)
   - `package.json` (`repository.url`)
   - Locale files (user-facing strings with GitHub links)
   - `README.BB` (BBCode version of README for WoWInterface)
   - Python scripts (`.scripts/missing_translation_check.py` output text)
   - Go const in `updatebin.go` (self-update target repo)

5. **GitHub redirects are indefinite** but break if someone creates a repo with the old name under the old owner. Treat them as a convenience bridge, not a permanent solution.

6. **Org-level rulesets require GitHub Team plan.** Free org plans only support repo-level rulesets. Workaround: maintain canonical config in `.github` org repo and apply via script.

7. **Reusable workflows should live in `wow-build-tools`, not `.github`.** Initially placed in `.github` org repo during Session 2. Later reconsidered: `wow-build-tools` is better because (a) external addon devs already use WBT and can leverage the same workflows, (b) WBT has versioned tags (`@v1-beta`) for stable pinning vs `.github`'s `@main`, and (c) Phase 6 scaffolding (`wow-build-tools init`) can generate callers that reference the same repo. **Action for next session:** move the 3 completed workflows from `.github` to `wow-build-tools` and update caller refs.

8. **`secrets: inherit` covers both org-level and repo-level secrets.** No need for explicit secret inputs as long as secret names are standardized across repos. This enabled standardizing Discord webhook secrets to `DISCORD_RELEASES_WEBHOOK` (same name, different values per repo).

9. **TOC files are a rich source of addon metadata.** `X-Curse-Project-ID`, `X-WoWI-ID`, `X-Wago-ID` can be parsed at workflow runtime to dynamically build distribution links, reducing inputs and preventing stale URLs.

10. **`gh pr edit` fails with a GraphQL deprecation warning about Projects Classic.** Workaround: use the REST API directly via `gh api -X PATCH repos/{owner}/{repo}/pulls/{number}`.

---

## Session 5: What Was Accomplished (March 3, 2026)

### Bugfixes

1. **toc-updater `startup_failure`** — All caller workflows with `permissions: {}` at top level were capping GITHUB_TOKEN to `none`, failing reusable workflows at startup validation. Fixed by adding job-level permissions (union of sub-job permissions) to all callers. Also moved `toc-updater.yml` permissions from workflow level to job level.
   - [Endeavoring PR #17](https://github.com/McTalian-WoW-Addons/Endeavoring/pull/17) ✅ Merged
   - [RPGLootFeed PR #530](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/530) ✅ Merged
   - [wow-build-tools PR #70](https://github.com/McTalian-WoW-Addons/wow-build-tools/pull/70) ✅ Merged

2. **WBT distro archive folder wrapping** — `zip -j` placed all files at archive root; CurseForge requires a top-level folder. Fixed `release-published.yml` to stage files into `wow-build-tools/` dir before zipping. Included in PR #70.

3. **`apply-ruleset.sh --all` exclusions** — Added `EXCLUDED_REPOS` array (`wow-build-tools`, `.github`) so `--all` skips repos that don't use the shared reusable workflows.

4. **Phase 4 cleanup** — `feat/reusable-workflows` branch hard-reset to match `origin/beta` (squash merge had already landed all content).

### Phase 5: Centralize Node Tooling — In Progress

**Goal:** Zero Node footprint in addon repos. All semantic-release deps, `.releaserc.json`, and `commitlint.config.js` centralized in WBT.

**WBT `feat/phase-5-consolidation` ([PR #71](https://github.com/McTalian-WoW-Addons/wow-build-tools/pull/71) — pending merge):**

- Added `commitlint.config.js` and `.releaserc.json` to WBT root (Endeavoring's version as canonical base)
- Added `wbt-ref` input (default `v1-beta`) to `ci.yml` and `pr-checks.yml` for test branch overrides
- `ci.yml`: checks out WBT into `.wbt/`, `npm ci --prefix .wbt`, copies `.releaserc.json`, runs `npx --prefix .wbt semantic-release`
- `pr-checks.yml`: checks out WBT into `.wbt/`, uses `.wbt/.nvmrc`, `--config .wbt/commitlint.config.js`

**Addon `chore/remove-node-tooling` branches (both repos):**

- Removed `package.json`, `package-lock.json`, `.releaserc.json`, `commitlint.config.js`, `.nvmrc`
- Temporarily calling WBT at `feat/phase-5-consolidation` with `wbt-ref` for testing
- [RPGLootFeed PR #532](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/532) — commitlint ✅ passed
- Endeavoring branch pushed, PR not yet opened

**Testing status:** commitlint passing ✅. Semantic-release path untested (needs push to main).

---

## What Remains

### Immediate next session

1. **Merge WBT PR #71** to beta, update `v1-beta` tag
2. **Update addon `chore/remove-node-tooling` branches:**
   - Swap `feat/phase-5-consolidation` → `v1-beta` in the `uses:` lines
   - Remove `wbt-ref: feat/phase-5-consolidation` inputs (default kicks in)
3. **Open Endeavoring PR**, merge both addon PRs, verify semantic-release runs cleanly on next push to main
4. **npm caching improvement:** Move `commitlint` devDependencies into WBT `package.json` so `npm ci --prefix .wbt` is cacheable (currently `npm install` with no lockfile in commitlint job)
5. **Lua version alignment:** RPGLootFeed CI still on 5.3.5 vs Endeavoring's 5.4.4

### Phase 3 remaining

- [ ] Retire old personal PAT (after verifying all workflows and migrating remaining addons)

### Phase 6: `wow-build-tools init` Command

Future. Scaffolding command for new addons.

### Future Addon Migrations

~3-5 more addons to migrate to the org. The same reference update sweep plus the thin caller workflow setup will be needed for each.

---

## Key Files & Locations

| Resource                             | Location                                                                                                               |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| **Migration plan** (with checkboxes) | `wow-build-tools/docs/org-migration-plan.md`                                                                           |
| **This session summary**             | `wow-build-tools/docs/org-migration-session-summary.md`                                                                |
| **Ruleset canonical config**         | `.github/rulesets/default.json`                                                                                        |
| **Ruleset apply script**             | `.github/scripts/apply-ruleset.sh`                                                                                     |
| **Reusable workflows**               | `wow-build-tools/.github/workflows/` (on `beta`, tagged `v1-beta`)                                                     |
| **WBT Phase 5 PR**                   | [PR #71](https://github.com/McTalian-WoW-Addons/wow-build-tools/pull/71) — `feat/phase-5-consolidation`, pending merge |
| **Endeavoring cleanup branch**       | `chore/remove-node-tooling` — PR not yet opened                                                                        |
| **RPGLootFeed cleanup PR**           | [PR #532](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/532) — `chore/remove-node-tooling`, commitlint ✅    |

---

## Reference: Files Intentionally Skipped

These were intentionally not updated during the reference sweep:

| Category                                                             | Why                                                               |
| -------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Go module path + imports (`github.com/McTalian/wow-build-tools/...`) | Works via redirect; planned for later (large scope, all Go files) |
| `.release/` directories                                              | Build outputs, regenerated on next build                          |
| `.scripts/.output/` markdown reports                                 | Generated by Python i18n scripts                                  |
| `org-migration-plan.md` reference table                              | Historical documentation showing old values                       |
| "McTalian" as a person/BattleTag in prose                            | Not a repo reference (e.g., sync-protocol.md examples, glossary)  |
