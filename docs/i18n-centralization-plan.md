# i18n Centralization Plan

**Created:** March 6, 2026  
**Status:** Phase 1 complete — Phase 2/3 ready to begin

The goal is to move all i18n tooling into `wow-build-tools` so that any addon can opt into `i18n-enabled: true` in the reusable CI/PR workflows without needing to maintain addon-specific Python scripts.

---

## Background

Currently `RPGLootFeed` and `BeaconUnitFrames` each maintain near-identical copies of four Python scripts under `.scripts/`. The reusable `ci.yml` and `pr-checks.yml` workflows already have an `i18n-enabled` flag, but it blindly calls `uv run .scripts/<script>.py` in the calling repo — meaning only addons that ship those scripts can use it.

**Addons and their i18n state:**

| Addon                   | Has locales | Has scripts | i18n-enabled in CI        |
| ----------------------- | ----------- | ----------- | ------------------------- |
| `RPGLootFeed`           | ✅          | ✅          | ✅                        |
| `BeaconUnitFrames`      | ✅          | ✅          | ❌ (blocked)              |
| `DeviceLayoutPreset`    | ✅          | ❌          | ❌ (blocked)              |
| `Endeavoring`           | ❌ (ticket) | ❌          | ❌ (blocked)              |
| `TokenTransmogTooltips` | ❌          | ❌          | N/A (no user-facing text) |

---

## Scripts Being Centralized

Five scripts, all shipping to `scripts/i18n/` in WBT:

| Script                             | What it does                                                                                          | Key params needed                                                                              |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `missing_translation_check.py`     | Compares each locale against enUS, outputs per-locale missing-key markdown reports for issue creation | `--locale-dir`, `--locale-xml`, `--ignored-files`, `--repo` (or read from `GITHUB_REPOSITORY`) |
| `create_or_update_i18n_issues.py`  | Creates/updates GitHub issues for each locale with missing translation report                         | reads `GITHUB_REPOSITORY` env var — no arg needed                                              |
| `check_for_missing_locale_keys.py` | Verifies all `ns.L["key"]` call sites in addon code have a matching definition in `enUS.lua`          | `--addon-dir`, `--locale-dir`, `--key-prefix` (default `L`, BUF uses `ns.L`)                   |
| `hardcode_string_check.py`         | Scans `.lua` files for hard-coded strings that should be localized                                    | `--addon-dir`, `--ignore-files`, `--ignore-dirs`                                               |
| `organize_translations.py`         | Sorts locale file entries within `--#region` blocks                                                   | `--locale-dir`, `--locale-xml`                                                                 |

**BUF improvements to bring forward (better than RLF versions):**

- `missing_translation_check.py`: improved regex handles single-quoted Lua string values
- `*`: `index.xml` as XML filename (addons can override), expanded `ignored_files` defaults

---

## Phase 1 — WBT: Ship generic scripts

**Branch:** `feat/centralize-i18n`  
**Status:** [x] Complete

### Tasks

- [x] Create `scripts/i18n/` directory in WBT
- [x] Port each script from BUF (as the more advanced base), replacing all hardcoded values with `argparse` CLI parameters
  - [x] `missing_translation_check.py`
  - [x] `create_or_update_i18n_issues.py`
  - [x] `check_for_missing_locale_keys.py`
  - [x] `hardcode_string_check.py`
  - [x] `organize_translations.py`
- [x] Add `scripts/i18n/pyproject.toml` with script deps: `defusedxml`, `requests`
- [x] Update `ci.yml` reusable workflow:
  - [x] Add inputs: `locale-dir` (default `{addon-name}/locale`), `locale-xml` (default `index.xml`), `locale-key-prefix` (default `L`)
  - [x] Checkout WBT repo into `.wbt/` at the pinned ref used by the calling workflow
  - [x] Run scripts from `.wbt/scripts/i18n/` via `uv run --with <dep> --no-project`
  - [x] Remove assumption that `.scripts/` exists in calling repo
- [x] Update `pr-checks.yml` reusable workflow with the same changes
- [x] Keep backward compatibility: if `.scripts/missing_translation_check.py` still exists in the calling repo, emit a `::warning::` annotation
- [x] Document inputs in workflow comments

---

## Phase 2 — BeaconUnitFrames: Migrate to centralized scripts

**Branch:** `chore/centralize-i18n`  
**Status:** [ ] Blocked on Phase 1

### Tasks

- [ ] Set `i18n-enabled: true` in `main.yml`
- [ ] Set `i18n-enabled: true` in `pr-checks.yml`
- [ ] Pass any non-default inputs (e.g. `locale-key-prefix: ns.L`, `locale-xml: index.xml`)
- [ ] Delete `.scripts/` (all 5 scripts now in WBT)
- [ ] Simplify or remove `pyproject.toml` (no longer needed for scripts; keep if used for other tooling)
- [ ] Verify CI passes

---

## Phase 3 — DeviceLayoutPreset: Enable i18n

**Branch:** `chore/enable-i18n`  
**Status:** [ ] Blocked on Phase 1

### Tasks

- [ ] Set `i18n-enabled: true` in `main.yml`
- [ ] Set `i18n-enabled: true` in `pr-checks.yml`
- [ ] Verify locale structure matches what the scripts expect (enUS.lua as reference, index.xml listing others)
- [ ] Add `pyproject.toml` (if uv needs it in the calling repo for lock file; TBD based on Phase 1 design)
- [ ] Verify CI passes

---

## Phase 4 — Endeavoring: Add locale support

**Branch:** `feat/locale-support`  
**Status:** [ ] Blocked on Phase 1; existing ticket for locale work

### Tasks

- [ ] Add `Endeavoring/locale/` directory
- [ ] Create `enUS.lua` with all current user-facing strings
- [ ] Create stub files for each supported locale (deDE, esES, esMX, frFR, itIT, koKR, ptBR, ruRU, zhCN, zhTW)
- [ ] Create `index.xml` listing all locale files
- [ ] Wire locale loading into `Bootstrap.lua` or `Core.lua`
- [ ] Set `i18n-enabled: true` in `main.yml` and `pr-checks.yml`
- [ ] Verify CI passes

---

## Phase 5 — RPGLootFeed: Migrate to centralized scripts

**Branch:** (defer — currently in active rearch on `feature-module-rearch`)  
**Status:** [ ] Blocked on Phase 1; deferred until rearch is merged

### Tasks

- [ ] Set non-default inputs in `main.yml`/`pr-checks.yml` if needed (e.g. `locale-xml: locales.xml`)
- [ ] Delete `.scripts/` i18n scripts (keep `post-pkg-comment.cjs`, `get_wowhead_hidden_currencies.py`, `check_for_invalid_prints.py` — those are RLF-specific)
- [ ] Simplify `pyproject.toml`
- [ ] Verify CI passes

---

## Future: `wow-build-tools init`

When `init` is implemented, i18n support should be a first-class option:

```bash
wow-build-tools init --name MyAddon --i18n
```

This would scaffold the `locale/` directory, stub locale files, `index.xml`, and set `i18n-enabled: true` in generated workflows from the start.

---

## Notes

- The BUF scripts are the more advanced baseline — use them as the canonical starting point, not the RLF versions
- `create_or_update_i18n_issues.py` reads `GITHUB_REPOSITORY` env var (set automatically by GitHub Actions) — no `--repo` arg needed
- `organize_translations.py` is not currently called by any reusable workflow step — it may be a developer utility rather than a CI script; clarify before wiring it up
- The `uv` Python environment setup is already in the reusable workflows (`astral-sh/setup-uv`); scripts just need to be pointed at the WBT checkout
