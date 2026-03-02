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

7. **Reusable workflows belong in `.github` org repo, not `wow-build-tools`.** The original plan placed them in `wow-build-tools`, but `workflow_call` workflows must live in a repo that callers can reference via `uses: org/repo/.github/workflows/file.yml@ref`. The `.github` org repo is the canonical home since it's already the org-wide shared config repo.

8. **`secrets: inherit` covers both org-level and repo-level secrets.** No need for explicit secret inputs as long as secret names are standardized across repos. This enabled standardizing Discord webhook secrets to `DISCORD_RELEASES_WEBHOOK` (same name, different values per repo).

9. **TOC files are a rich source of addon metadata.** `X-Curse-Project-ID`, `X-WoWI-ID`, `X-Wago-ID` can be parsed at workflow runtime to dynamically build distribution links, reducing inputs and preventing stale URLs.

10. **`gh pr edit` fails with a GraphQL deprecation warning about Projects Classic.** Workaround: use the REST API directly via `gh api -X PATCH repos/{owner}/{repo}/pulls/{number}`.

---

## What Remains

### Phase 3: Finish Reference Updates ✅

- [x] Open PR for Endeavoring `update-more-references` branch and merge
- [x] Merge RPGLootFeed PR #527
- [x] Open PR for wow-build-tools `chore/update-org-references` branch and merge
- [ ] Retire old personal PAT (Phase 2, step 2a.3 — after verifying all workflows and migrating remaining addons)

### Phase 3.5: Standardize Repo Rulesets ✅

- [x] Audit existing rulesets across all 3 repos
- [x] Design standardized config
- [x] Apply via API to RPGLootFeed, Endeavoring, wow-build-tools
- [x] Create reusable script + canonical JSON in `.github` org repo
- [x] Push to `.github` main

### Phase 4: Create Reusable Workflows — In Progress

Reusable workflows live in `McTalian-WoW-Addons/.github/.github/workflows/` (not wow-build-tools as originally planned — `.github` org repo is the correct home for `workflow_call` workflows).

#### Completed (Session 2, March 2)

1. **`cleanup-stale-issues.yml`** ✅ — Zero inputs, identical config. Proved the pattern.
2. **`toc-updater.yml`** ✅ — Inputs: `addon-name`, `pr-body`. RPGLootFeed's hidden currencies steps split into a separate `hidden-currencies.yml` workflow (weekly Monday 2pm UTC).
3. **`package-and-distribute.yml`** ✅ — Inputs: `addon-name`, `avatar-url`. Distribution links (CurseForge, WoWInterface, Wago) parsed dynamically from the addon's `.toc` file (`X-Curse-Project-ID`, `X-WoWI-ID`, `X-Wago-ID`). Missing IDs → that link is omitted from the Discord announcement.
   - **Discord webhook secrets standardized:** `DISCO_WH_NDVRNG_RELEASES` → `DISCORD_RELEASES_WEBHOOK` (Endeavoring), `DISCO_WH_RLF_RELEASES` → `DISCORD_RELEASES_WEBHOOK` (RPGLootFeed). Same name, different values per repo.

**PRs open (both repos, same branch `chore/reusable-stale-workflow`):**

- [Endeavoring PR #15](https://github.com/McTalian-WoW-Addons/Endeavoring/pull/15) — stale + toc-updater + package-and-distribute
- [RPGLootFeed PR #528](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/528) — stale + toc-updater + hidden-currencies + package-and-distribute

#### Remaining

4. `ci.yml` / `main.yml` (parameterized with addon name, Lua version, i18n toggle)
5. `pr-checks.yml` (parameterized, needs `post-pkg-comment.cjs` consolidation first)

### Phase 5: Consolidate Scripts & Config

Not started. Key items:

- **`post-pkg-comment.cjs`** — Endeavoring version is more robust; backport to RPGLootFeed, then move to wow-build-tools
- **Identical files to share:** `commitlint.config.js`, `.releaserc.json`
- **Align `.nvmrc`** (`v24` vs `24.10.0`)
- **Align Lua version** (Endeavoring CI: 5.4, RPGLootFeed CI: 5.3 — both rockspecs say `>= 5.3`)

### Phase 6: `wow-build-tools init` Command

Future. Scaffolding command for new addons.

### Future Addon Migrations

~3-5 more addons to migrate to the org. Keep in mind:

- When `package.json` is dropped (Phase 5), semantic-release will need `repositoryUrl` in `.releaserc.json` or will use CI-detected URL.
- The same reference update sweep will be needed for each addon.

---

## Key Files & Locations

| Resource                              | Location                                                                                                  |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| **Migration plan** (with checkboxes)  | `wow-build-tools/docs/org-migration-plan.md`                                                              |
| **This session summary**              | `wow-build-tools/docs/org-migration-session-summary.md`                                                   |
| **Endeavoring pending branch**        | Merged                                                                                                    |
| **RPGLootFeed reference PR**          | [PR #527](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/527) — Merged                           |
| **wow-build-tools pending branch**    | Merged                                                                                                    |
| **Ruleset canonical config**          | `.github/rulesets/default.json`                                                                           |
| **Ruleset apply script**              | `.github/scripts/apply-ruleset.sh`                                                                        |
| **Reusable workflows**                | `.github/.github/workflows/` (cleanup-stale-issues, toc-updater, package-and-distribute)                  |
| **Endeavoring reusable workflows PR** | [PR #15](https://github.com/McTalian-WoW-Addons/Endeavoring/pull/15) on `chore/reusable-stale-workflow`   |
| **RPGLootFeed reusable workflows PR** | [PR #528](https://github.com/McTalian-WoW-Addons/RPGLootFeed/pull/528) on `chore/reusable-stale-workflow` |

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
