# Future Work

**Last updated:** March 6, 2026

Captures remaining tasks and future improvements across the `McTalian-WoW-Addons` org.

---

## Immediate: Remaining Addon Migrations to Org

The following addons still live under the personal `McTalian` account and need transferring to `McTalian-WoW-Addons`. The old personal PAT cannot be retired until these are migrated.

| Addon                   | Priority | Notes                       |
| ----------------------- | -------- | --------------------------- |
| `DeviceLayoutPreset`    | High     | ✅ Migrated (March 5, 2026) |
| `BeaconUnitFrames`      | High     | ✅ Migrated (March 5, 2026) |
| `TokenTransmogTooltips` | High     | ✅ Migrated (March 5, 2026) |
| `LibPixelPerfect-1.0`   | Low      | Library, lower urgency      |

**For each addon, the migration checklist is:**

- [ ] Transfer repo to `McTalian-WoW-Addons`
- [ ] Update local git remote
- [ ] Delete repo-level secrets now covered by org-level (`GH_PAT`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`)
- [ ] Add repo-level secret `DISCORD_RELEASES_WEBHOOK`
- [ ] Re-add branch protection bypass as org owner/admin (bypass lists don't transfer)
- [ ] Update all `McTalian/` references to `McTalian-WoW-Addons/` in workflows, TOC, README, package.json, etc.
- [ ] Replace inline workflows with thin callers to WBT reusable workflows
- [ ] Delete `package.json` and `package-lock.json` (semantic-release now handled by reusable CI workflow)
- [ ] Verify CI passes with org-level secrets

**After all addons are migrated:**

- [x] Retire old personal PAT (`McTalian` account fine-grained token)

---

## WBT: `wow-build-tools init` Command

Scaffolding subcommand for onboarding new addons quickly.

```bash
wow-build-tools init --name MyAddon --flavors retail,classic --platforms curseforge,wago,wowi
```

**Would generate:**

- Thin caller workflows (`.github/workflows/`) pointing to WBT reusable workflows
- `renovate.json` (see `wow-build-tools/renovate.json` for the current template)
- `Makefile` (parameterized)
- Rockspec
- `.luacov`
- Basic addon skeleton (`MyAddon/MyAddon.toc`, `MyAddon/Core.lua`)

When `--i18n` is passed, also scaffold:

- `locale/` directory with `enUS.lua` stub and locale stubs for all supported locales
- `index.xml` listing all locale files
- Wire `i18n-enabled: true` in generated workflows

- [ ] Design the command interface and template system
- [ ] Implement template files
- [ ] Add to wow-build-tools CLI

---

## ~~WBT: Centralize i18n Tooling for Reusable Workflows~~ ✅ Complete (March 2026)

All 5 phases complete. Generic, parameterized i18n scripts now ship in `scripts/i18n/` and are checked out into `.wbt/` by the reusable `ci.yml` and `pr-checks.yml` workflows. All addons have been migrated.

| Addon                 | Status                                      |
| --------------------- | ------------------------------------------- |
| RPGLootFeed           | ✅ Migrated (Phase 5)                       |
| BeaconUnitFrames      | ✅ Enabled (Phase 2)                        |
| DeviceLayoutPreset    | ✅ Enabled (Phase 3)                        |
| Endeavoring           | ✅ Locale support added + enabled (Phase 4) |
| TokenTransmogTooltips | N/A (no user-facing text)                   |

---

## WBT: Pre-Release Checklist Command

Comprehensive pre-release validation to run before tagging a version or releasing.

**Potential checks:**

- TOC file validation (version, dependencies, load order)
- SavedVariables integrity checks
- Error handling audit
- Documentation completeness check
- Breaking change identification

- [ ] Design the command interface
- [ ] Implement checks

---

## Tooling: Pin GitHub Actions to Commit SHAs (`pinact`)

Run `pinact` as a one-shot local operation to replace mutable tag refs (e.g. `@v4`, `@v1`) with pinned commit SHAs in all workflow files. Renovate (`github-actions` datasource, grouped under the `github-actions` packageRule) will keep them updated from there.

**Repos to run against:** `Endeavoring`, `RPGLootFeed`, `wow-build-tools`, `DeviceLayoutPreset`

- [ ] Run `pinact run` across all four repos
- [ ] Commit and push the pinned refs

---

## Low Priority: Minor Config Alignment

Small inconsistencies between addon repos, not worth a dedicated sprint but worth cleaning up when touching these files.

| File                    | Status                                         |
| ----------------------- | ---------------------------------------------- |
| `.luacov`               | Minor differences between repos — low priority |
| Rockspec                | Minor differences — low priority               |
| Makefile (core targets) | Minor differences — low priority               |

---

## WoW Forever (1.60.x) — remaining unknowns

Forever beta opened 2026-09-17 (ends 2026-10-21); launch 2026-11-04. Classification
support landed already: `Forever` game flavor, interface range `16xxx`
(1.60.1 → `16001`), a minor-version split against Classic Era, and a guard in
`CheckForInterfaceBumps` that rejects a product serving a different client line
(`wow_classic_beta` is currently serving the Forever beta, not Mists).

These are guesses marked `TODO(forever)` in the code and must be confirmed:

- [ ] Install directory names (`_forever_`, `_forever_beta_`) — `internal/flavor/flavor.go`
- [ ] CDN product codes (`wow_forever`, `wow_forever_beta`, `wow_forever_ptr`) — `internal/toc/interface_versions.go`
- [ ] TOC filename/`## Interface-` suffix (`Forever`) — `internal/toc/general.go`
- [ ] CurseForge `gameVersionTypeID` and whether `1.60.x` appears in `/api/game/wow/versions`
- [ ] Wago `patches` key and `toc_suffixes` entry (absent as of 2026-09-16)
- [ ] WoWInterface compatibility id (absent as of 2026-09-16; they also still lack Mists and Titan)
- [ ] `.release.json` flavor string — currently emitted as `forever`

**Cross-repo:** `toc-interface-updater` has the same guard
(`checked_product_version` in `toc_interface_updater/update.py`) plus a `forever`
flavor, a `_Forever.toc` suffix and `## Interface-Forever` handling. Its three
live tests need a machine that can reach `us.version.battle.net:1119` — they skip
when the version server is unreachable.

**Unrelated pre-existing gap:** Titan (`3.80.x`) has no flavor here. BigWigs
packager has `380??` → `titan` and CurseForge `81212`; Wago publishes a `titan`
patches key.
