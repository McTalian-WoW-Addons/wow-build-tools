# Organization Migration Plan: `McTalian-WoW-Addons`

**Created:** February 28, 2026
**Status:** In Progress

This plan covers migrating `Endeavoring`, `RPGLootFeed`, and `wow-build-tools` to the `McTalian-WoW-Addons` GitHub organization, consolidating shared tooling, and creating reusable workflows.

---

## Table of Contents

- [Phase 1: Pre-Transfer Preparation](#phase-1-pre-transfer-preparation)
- [Phase 2: Transfer Repos](#phase-2-transfer-repos)
- [Phase 3: Update Hardcoded References](#phase-3-update-hardcoded-references)
- [Phase 4: Create Reusable Workflows](#phase-4-create-reusable-workflows)
- [Phase 5: Consolidate Supporting Scripts & Config](#phase-5-consolidate-supporting-scripts--config)
- [Phase 6: Future — `wow-build-tools init` Command](#phase-6-future--wow-build-tools-init-command)
- [Reference: Secrets & Variables Inventory](#reference-secrets--variables-inventory)
- [Reference: Hardcoded References Inventory](#reference-hardcoded-references-inventory)
- [Reference: Shared vs Addon-Specific Files](#reference-shared-vs-addon-specific-files)

---

## Phase 1: Pre-Transfer Preparation

> Set up the org so everything works seamlessly post-transfer.

### 1a. Configure Org-Level Secrets

Set these at **Organization Settings → Secrets and variables → Actions → Organization secrets**, scoped to all repos:

- [x] `CF_API_KEY` — CurseForge uploads
- [x] `WOWI_API_TOKEN` — WoWInterface uploads
- [x] `WAGO_API_TOKEN` — Wago.io uploads
- [x] `GH_PAT` — Semantic-release, PR creation, cross-repo actions

### 1b. Keep These as Repo-Level Secrets

These are inherently per-addon and stay at the repo level:

| Secret                     | Repo        | Purpose                       |
| -------------------------- | ----------- | ----------------------------- |
| `DISCORD_RELEASES_WEBHOOK` | Endeavoring | Discord release announcements |
| `DISCORD_RELEASES_WEBHOOK` | RPGLootFeed | Discord release announcements |

> **Note:** Originally named `DISCO_WH_NDVRNG_RELEASES` and `DISCO_WH_RLF_RELEASES`. Renamed to standardized `DISCORD_RELEASES_WEBHOOK` during Phase 4 reusable workflow work (March 2, 2026).

### 1c. Create Org `.github` Repo (Optional, Low Priority)

A `McTalian-WoW-Addons/.github` repo can hold:

- [x] Org profile README
- [ ] Default issue/PR templates
- [ ] Default CONTRIBUTING.md, CODE_OF_CONDUCT.md

---

## Phase 2: Transfer Repos

Use GitHub's **Transfer repository** feature (Settings → Danger Zone → Transfer).

### 2a. Re-issue `GH_PAT` for the Org (Do Before Transfers)

The existing fine-grained PAT is scoped to `McTalian` (personal account) with selected repo access. After transfer, the repos live under `McTalian-WoW-Addons` and the old token won't have access.

**Steps:**

1. - [x] Create a **new fine-grained PAT** at [github.com/settings/personal-access-tokens/new](https://github.com/settings/personal-access-tokens/new):
   - **Resource owner:** `McTalian-WoW-Addons`
   - **Repository access:** All repositories (or select the three being transferred)
   - **Permissions:** `contents: write`, `pull-requests: write`, `issues: write` (match current token)
2. - [x] Update the **org-level `GH_PAT` secret** with the new token value
3. - [ ] Retire/revoke the old personal PAT (after verifying workflows work post-transfer)

> **Why:** `GH_PAT` is used by semantic-release (creating releases), TOC updater (creating PRs), and wow-build-tools CI. All will fail post-transfer without an org-scoped token.

### 2b. Transfer Order

Transfer `wow-build-tools` first since the addons reference it:

1. - [x] **`wow-build-tools`** → `McTalian-WoW-Addons/wow-build-tools`
   - [x] Update local git remote: `git remote set-url origin git@github.com:McTalian-WoW-Addons/wow-build-tools.git`
   - [x] Delete repo-level secrets now covered by org-level secrets (`GH_PAT`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`)
   - [x] Verify workflows run with org-level secrets
   - [x] Verify branch protection rules transferred
2. - [x] **`Endeavoring`** → `McTalian-WoW-Addons/Endeavoring`
   - [x] Update local git remote: `git remote set-url origin git@github.com:McTalian-WoW-Addons/Endeavoring.git`
   - [x] Delete repo-level secrets now covered by org-level secrets (`GH_PAT`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`)
   - [ ] Verify workflows run with org-level secrets
   - [x] Verify branch protection rules transferred (check bypass list — see note below)
3. - [x] **`RPGLootFeed`** → `McTalian-WoW-Addons/RPGLootFeed`
   - [x] Update local git remote: `git remote set-url origin git@github.com:McTalian-WoW-Addons/RPGLootFeed.git`
   - [x] Delete repo-level secrets now covered by org-level secrets (`GH_PAT`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`)
   - [ ] Verify workflows run with org-level secrets
   - [x] Verify branch protection rules transferred (check bypass list — see note below)

### What Transfers Automatically

- All issues, PRs, stars, watchers, forks
- Branch protection rules — ⚠️ **bypass lists do NOT transfer**; re-add bypasses using org owners/admins instead of personal account admins
- GitHub Actions secrets (repo-level) — ⚠️ **these override org-level secrets with the same name, so delete them**
- Webhooks
- GitHub sets up redirects from `McTalian/<repo>` → `McTalian-WoW-Addons/<repo>`

---

## Phase 3: Update Hardcoded References

> Redirects work post-transfer but should be treated as temporary. Update all `McTalian/` references to `McTalian-WoW-Addons/`.

### Workflow Action References

| File                                      | Old Reference                                 | New Reference                                            |
| ----------------------------------------- | --------------------------------------------- | -------------------------------------------------------- |
| `Endeavoring: package-and-distribute.yml` | `McTalian/wow-build-tools@v1-beta`            | `McTalian-WoW-Addons/wow-build-tools@v1-beta`            |
| `Endeavoring: pr-checks.yml`              | `McTalian/wow-build-tools@v1-beta`            | `McTalian-WoW-Addons/wow-build-tools@v1-beta`            |
| `Endeavoring: toc-updater.yml`            | `Mctalian/wow-build-tools/toc/update@v1-beta` | `McTalian-WoW-Addons/wow-build-tools/toc/update@v1-beta` |
| `RPGLootFeed: package-and-distribute.yml` | `McTalian/wow-build-tools@v1-beta`            | `McTalian-WoW-Addons/wow-build-tools@v1-beta`            |
| `RPGLootFeed: pr-checks.yml`              | `McTalian/wow-build-tools@v1-beta`            | `McTalian-WoW-Addons/wow-build-tools@v1-beta`            |
| `RPGLootFeed: toc-updater.yml`            | `Mctalian/wow-build-tools/toc/update@v1-beta` | `McTalian-WoW-Addons/wow-build-tools/toc/update@v1-beta` |

### Discord Embed / Message URLs

| File                                                             | Old URL Fragment       | New URL Fragment                  |
| ---------------------------------------------------------------- | ---------------------- | --------------------------------- |
| `Endeavoring: package-and-distribute.yml` (avatar-url)           | `McTalian/Endeavoring` | `McTalian-WoW-Addons/Endeavoring` |
| `Endeavoring: package-and-distribute.yml` (GitHub links in body) | `McTalian/Endeavoring` | `McTalian-WoW-Addons/Endeavoring` |
| `RPGLootFeed: package-and-distribute.yml` (avatar-url)           | `McTalian/RPGLootFeed` | `McTalian-WoW-Addons/RPGLootFeed` |
| `RPGLootFeed: package-and-distribute.yml` (GitHub links in body) | `McTalian/RPGLootFeed` | `McTalian-WoW-Addons/RPGLootFeed` |

### wow-build-tools Internal References

| File                    | Old Reference             | New Reference                        |
| ----------------------- | ------------------------- | ------------------------------------ |
| `release-published.yml` | `McTalian/${{ env.CMD }}` | `McTalian-WoW-Addons/${{ env.CMD }}` |

### Progress

- [x] Endeavoring: All references updated
- [x] RPGLootFeed: All references updated
- [x] wow-build-tools: All references updated

---

## Phase 4: Create Reusable Workflows

> Convert duplicated workflows into parameterized [reusable workflows](https://docs.github.com/en/actions/sharing-automations/reusing-workflows) in `McTalian-WoW-Addons/.github`.
>
> **Note:** Originally planned for `wow-build-tools`, but `.github` org repo is the correct home for `workflow_call` workflows.

### 4a. `cleanup-stale-issues.yml` — Zero Inputs ✅

**Effort:** ~30 min | **Priority:** First — proves the pattern

100% identical between both repos. No inputs needed.

```yaml
# Caller workflow in each addon repo (entire file):
name: Close stale issues
on:
  schedule:
    - cron: 30 1 * * *
permissions: {}
jobs:
  stale:
    uses: McTalian-WoW-Addons/.github/.github/workflows/cleanup-stale-issues.yml@main
```

- [x] Create reusable workflow in `.github` org repo
- [x] Convert Endeavoring caller
- [x] Convert RPGLootFeed caller

### 4b. `package-and-distribute.yml` ✅

**Effort:** ~45 min | **Priority:** High

**Inputs:**

- `addon-name` — addon directory name
- `avatar-url` — Discord embed avatar image URL

**Dynamic from TOC file:**

- `X-Curse-Project-ID` → CurseForge link (slug = lowercased addon name)
- `X-WoWI-ID` → WoWInterface link
- `X-Wago-ID` → Wago link
- Missing IDs → that distribution link is omitted from Discord announcement

**Secrets:** `secrets: inherit` from caller (pulls org + repo secrets). Discord webhook standardized to `DISCORD_RELEASES_WEBHOOK` (repo-level, same name in every repo).

- [x] Standardize Discord webhook secret names across repos
- [x] Create reusable workflow in `.github` org repo
- [x] Convert Endeavoring caller
- [x] Convert RPGLootFeed caller

### 4c. `ci.yml` (replaces `main.yml`)

**Effort:** ~1 hour | **Priority:** High

**Inputs:**

- `addon-name`
- `rockspec-name`
- `lua-version` (e.g., `"5.4"`)
- `has-i18n` (boolean) — controls whether i18n translation job runs
- `spec-dir` — e.g., `Endeavoring_spec` or `RPGLootFeed_spec`

- [ ] Create reusable workflow in `.github` org repo
- [ ] Convert Endeavoring caller
- [ ] Convert RPGLootFeed caller
- [ ] Verify tests + semantic-release work on both repos

### 4d. `pr-checks.yml`

**Effort:** ~1.5 hours | **Priority:** High

**Inputs:**

- Same as ci.yml, plus:
- `has-trunk` (boolean) — controls trunk linting job
- `has-i18n` (boolean) — controls translation checks

**Note:** `post-pkg-comment.cjs` should be consolidated first (see Phase 5) or embedded into the reusable workflow.

- [ ] Consolidate `post-pkg-comment.cjs` (see Phase 5)
- [ ] Create reusable workflow in `.github` org repo
- [ ] Convert Endeavoring caller
- [ ] Convert RPGLootFeed caller
- [ ] Verify PR checks work on both repos

### 4e. `toc-updater.yml` ✅

**Effort:** ~45 min | **Priority:** Medium

**Inputs:**

- `addon-name`
- `pr-body` — PR description template (differs by addon due to flavor support)

**Design decision:** RPGLootFeed's hidden currency generation was split into a separate `hidden-currencies.yml` workflow (runs weekly on Monday 2pm UTC, creates its own PR branch). This keeps the reusable toc-updater clean and generic.

- [x] Create reusable workflow in `.github` org repo
- [x] Convert Endeavoring caller
- [x] Convert RPGLootFeed caller
- [x] Create separate `hidden-currencies.yml` for RPGLootFeed

---

## Phase 5: Consolidate Supporting Scripts & Config

### `post-pkg-comment.cjs`

The scripts **differ significantly** between the two addons. Endeavoring's version is more robust:

- Uses a formatted markdown table
- Handles first-release scenarios (no previous sizes to compare)
- Conditional standard/nolib package display
- `formatBytes()` utility for human-readable sizes

**Action:**

- [ ] Backport Endeavoring's improvements to RPGLootFeed
- [ ] Move the consolidated script into `wow-build-tools` (either as part of the reusable PR checks workflow or as a standalone composite action)

### Identical Config Files

These files are 100% identical between both addons and can be shared:

| File                   | Strategy                                                          |
| ---------------------- | ----------------------------------------------------------------- |
| `commitlint.config.js` | Publish as shared npm config or include in `wow-build-tools init` |
| `.releaserc.json`      | Same — shared config or template                                  |

- [ ] Decide on sharing mechanism (npm package vs template vs reusable workflow artifact)

### Near-Identical Config Files (Parameterize)

| File                    | Differs Only By                             |
| ----------------------- | ------------------------------------------- |
| `package.json`          | Addon name, description, keywords, repo URL |
| `.luacov`               | Project name, include pattern, exclude list |
| Rockspec                | Package name, source URL                    |
| Makefile (core targets) | Addon name, spec dir, rockspec name         |
| `.nvmrc`                | `v24` vs `24.10.0`                          |

- [ ] Standardize `.nvmrc` to a single format
- [ ] Consider including these in `wow-build-tools init` scaffolding (Phase 6)

### Lua Version Alignment

**Current state:** Endeavoring CI uses Lua 5.4.4, RPGLootFeed CI uses Lua 5.3.5. Both rockspecs say `>= 5.3`. Both Makefiles install with `--lua-version 5.4`.

- [ ] Align on Lua 5.4 across both addon CI pipelines

---

## Phase 6: Future — `wow-build-tools init` Command

**Effort:** ~Half day | **Priority:** Low (nice-to-have for future addon onboarding)

Add a scaffolding subcommand to `wow-build-tools`:

```bash
wow-build-tools init --name MyAddon --flavors retail,classic --platforms curseforge,wago,wowi
```

**Would generate:**

- Makefile (parameterized with addon name)
- Rockspec
- `.luacov`
- `package.json` + `.releaserc.json` + `commitlint.config.js`
- `.nvmrc`
- Thin caller workflows (`.github/workflows/`)
- Basic addon skeleton (`MyAddon/MyAddon.toc`, `MyAddon/Core.lua`)

- [ ] Design the command interface and template system
- [ ] Implement template files
- [ ] Add to wow-build-tools CLI

---

## Execution Timeline

| Phase                                        | Effort     | Status         |
| -------------------------------------------- | ---------- | -------------- |
| **Phase 1**: Org secrets                     | ~10 min    | ✅ Complete    |
| **Phase 2**: Transfer repos                  | ~15 min    | ✅ Complete    |
| **Phase 3**: Update references               | ~30 min    | ✅ Complete    |
| **Phase 3.5**: Standardize rulesets          | ~1 hour    | ✅ Complete    |
| **Phase 4a**: Stale issues reusable workflow | ~30 min    | ✅ Complete    |
| **Phase 4b**: Package & distribute reusable  | ~45 min    | ✅ Complete    |
| **Phase 4c**: CI reusable workflow           | ~1 hour    | ⬜ Not started |
| **Phase 4d**: PR checks reusable workflow    | ~1.5 hours | ⬜ Not started |
| **Phase 4e**: TOC updater reusable workflow  | ~45 min    | ✅ Complete    |
| **Phase 5**: Script/config consolidation     | ~1 hour    | ⬜ Not started |
| **Phase 6**: `init` command                  | ~Half day  | ⬜ Future      |

---

## Reference: Secrets & Variables Inventory

### Secrets Used Across Repos

| Secret                     | Endeavoring | RPGLootFeed | wow-build-tools |  Org-Level?   |
| -------------------------- | :---------: | :---------: | :-------------: | :-----------: |
| `CF_API_KEY`               |     ✅      |     ✅      |        —        |    ✅ Yes     |
| `WOWI_API_TOKEN`           |     ✅      |     ✅      |        —        |    ✅ Yes     |
| `WAGO_API_TOKEN`           |     ✅      |     ✅      |        —        |    ✅ Yes     |
| `GH_PAT`                   |     ✅      |     ✅      |       ✅        |    ✅ Yes     |
| `GITHUB_TOKEN`             |  ✅ (auto)  |  ✅ (auto)  |        —        | Auto-provided |
| `DISCORD_RELEASES_WEBHOOK` |     ✅      |     ✅      |        —        | ❌ Repo-level |

> **Note:** `DISCO_WH_NDVRNG_RELEASES` and `DISCO_WH_RLF_RELEASES` were renamed to `DISCORD_RELEASES_WEBHOOK` during Phase 4 (March 2, 2026).

### Variables

No `vars.*` references found in any workflow — everything currently uses secrets or hardcoded values.

---

## Reference: Hardcoded References Inventory

All occurrences of `McTalian/` in workflow files:

| File                                      | Line | Reference                                     |
| ----------------------------------------- | ---- | --------------------------------------------- |
| `Endeavoring: pr-checks.yml`              | 110  | `McTalian/wow-build-tools@v1-beta`            |
| `Endeavoring: package-and-distribute.yml` | 29   | `McTalian/wow-build-tools@v1-beta`            |
| `Endeavoring: package-and-distribute.yml` | 46   | `McTalian/Endeavoring` (avatar URL)           |
| `Endeavoring: package-and-distribute.yml` | 51   | `McTalian/Endeavoring` (releases link)        |
| `Endeavoring: package-and-distribute.yml` | 58   | `McTalian/Endeavoring` (issues link)          |
| `Endeavoring: toc-updater.yml`            | 23   | `Mctalian/wow-build-tools/toc/update@v1-beta` |
| `RPGLootFeed: package-and-distribute.yml` | 29   | `McTalian/wow-build-tools@v1-beta`            |
| `RPGLootFeed: package-and-distribute.yml` | 46   | `McTalian/RPGLootFeed` (avatar URL)           |
| `RPGLootFeed: package-and-distribute.yml` | 51   | `McTalian/RPGLootFeed` (releases link)        |
| `RPGLootFeed: package-and-distribute.yml` | 58   | `McTalian/RPGLootFeed` (issues link)          |
| `RPGLootFeed: toc-updater.yml`            | 23   | `Mctalian/wow-build-tools/toc/update@v1-beta` |
| `RPGLootFeed: pr-checks.yml`              | 96   | `McTalian/wow-build-tools@v1-beta`            |
| `wow-build-tools: release-published.yml`  | 237  | `McTalian/${{ env.CMD }}`                     |

---

## Reference: Shared vs Addon-Specific Files

| Component                    |     Shareable?      | What Differs                                   |
| ---------------------------- | :-----------------: | ---------------------------------------------- |
| `cleanup-stale-issues.yml`   |       ✅ 100%       | Nothing — identical                            |
| `commitlint.config.js`       |       ✅ 100%       | Nothing — identical                            |
| `.releaserc.json`            |       ✅ 100%       | Nothing — identical                            |
| `package-and-distribute.yml` |  📐 Templatizable   | Addon name, Discord webhook, branding          |
| `main.yml`                   |  📐 Templatizable   | Addon name, rockspec, Lua version, i18n toggle |
| `pr-checks.yml`              |  📐 Templatizable   | Addon name, Lua version, i18n/trunk toggles    |
| `toc-updater.yml`            |  📐 Templatizable   | Addon name, PR body, extra steps               |
| `package.json`               |  📐 Templatizable   | Metadata only                                  |
| Rockspec                     |  📐 Templatizable   | Package name, source URL                       |
| `.luacov`                    |  📐 Templatizable   | Project name, include/exclude patterns         |
| Makefile (core)              |  📐 Templatizable   | Addon name, spec dir                           |
| `.nvmrc`                     |    ⚠️ Misaligned    | `v24` vs `24.10.0`                             |
| Makefile (i18n/lint)         | 🔒 RPGLootFeed-only | Python scripts not in Endeavoring              |
| Python/uv tooling            | 🔒 RPGLootFeed-only | No pyproject.toml in Endeavoring               |
| Trunk config                 | 🔒 RPGLootFeed-only | No .trunk/ in Endeavoring                      |
| `post-pkg-comment.cjs`       |     ⚠️ Diverged     | Endeavoring version more robust                |
