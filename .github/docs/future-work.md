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

## WoW Forever (1.60.x)

Beta opened 2026-09-17 (ends 2026-10-21); launch 2026-11-04.

Confirmed as of 2026-09-17:

| Fact                         | Value                                      | Source                                                                      |
| ---------------------------- | ------------------------------------------ | --------------------------------------------------------------------------- |
| Interface range              | `16xxx`; 1.60.1 → `16001`                  | BigWigs packager `16???`; wow-ui-source `forever` branch `version.txt`      |
| Flavor slug                  | `forever` (alias `camelot`)                | packager `game_flavor` map                                                  |
| TOC suffix                   | `Camelot` (`_Camelot.toc`, `-Camelot.toc`) | packager globs and `forever) new_file+="_Camelot.toc"`                      |
| Build token                  | `@version-forever@`                        | packager lua/xml/toc filters                                                |
| CurseForge game version type | `88568`                                    | packager `forever) game_id=88568`                                           |
| Wago patches key             | `forever`                                  | `addons.wago.io/api/data/game`                                              |
| Install dir (beta)           | `_classic_beta_`                           | `wow_classic_beta` TACT product config `shared_container_default_subfolder` |

Install directories for every product come from that product's TACT product
config, which is the authoritative source:
`http://level3.blizzard.com/tpr/configs/data/<h0:2>/<h2:4>/<product_config>`,
field `all.config.shared_container_default_subfolder`. The `product_config`
hashes come from `https://wago.tools/api/builds/latest`.

Still open:

- [ ] Live CDN product code. None exists yet; the beta ships through
      `wow_classic_beta`. `wow_forever`, `wow_forever_beta` and `wow_forever_ptr`
      are declared on the assumption a standalone product appears before launch —
      unknown products are simply absent from the build feed, so they are inert.
- [ ] Live install directory. `_forever_` and `_forever_beta_` are declared for
      the same reason and do not exist yet; `Forever` also maps to the
      `classicBeta` install flavor so `link` works against the beta client today.
- [ ] Whether `1.60.x` appears in CurseForge `/api/game/wow/versions`. Needs a
      token; we match by version name rather than type id, so the id above is
      informational.
- [ ] WoWInterface support. Absent, and the packager explicitly warns and skips
      Forever uploads to WoWI. They also still lack Mists and Titan.

**Cross-repo:** `McTalian/wow-toc-updater` (formerly `toc-interface-updater`) is
unmaintained — every tag now fails with a migration notice pointing here.

---

## Flavor coverage

Confirmed install directories, from the TACT product configs described above:

| Product               | Version | Install dir         |
| --------------------- | ------- | ------------------- |
| `wow`                 | 12.1.0  | `_retail_`          |
| `wow_beta`            | 12.0.1  | `_beta_`            |
| `wowt`                | 12.1.0  | `_ptr_`             |
| `wowxptr`             | 12.1.5  | `_xptr_`            |
| `wow_classic`         | 5.5.4   | `_classic_`         |
| `wow_classic_ptr`     | 5.5.4   | `_classic_ptr_`     |
| `wow_classic_beta`    | 1.60.1  | `_classic_beta_`    |
| `wow_classic_era`     | 1.15.9  | `_classic_era_`     |
| `wow_classic_era_ptr` | 2.5.6   | `_classic_era_ptr_` |
| `wow_anniversary`     | 2.5.6   | `_anniversary_`     |
| `wow_classic_titan`   | 3.80.2  | `_classic_titan_`   |
| `wowlivetest`         | 10.2.5  | `_dark_realm_`      |
| `wowz`                | 1.14.4  | `_submission_`      |

Note that `wow_classic_beta` and `wow_classic_era_ptr` are both serving a
different client line than their name suggests. `CheckForInterfaceBumps` trusts
the build version over the product name for exactly this reason.
