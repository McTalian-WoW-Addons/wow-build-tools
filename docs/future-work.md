# Future Work

**Last updated:** March 4, 2026

Captures remaining tasks and future improvements across the `McTalian-WoW-Addons` org.

---

## Immediate: Remaining Addon Migrations to Org

The following addons still live under the personal `McTalian` account and need transferring to `McTalian-WoW-Addons`. The old personal PAT cannot be retired until these are migrated.

| Addon                   | Priority | Notes                  |
| ----------------------- | -------- | ---------------------- |
| `DeviceLayoutPreset`    | High     | —                      |
| `BeaconUnitFrames`      | High     | —                      |
| `TokenTransmogTooltips` | High     | —                      |
| `LibPixelPerfect-1.0`   | Low      | Library, lower urgency |

**For each addon, the migration checklist is:**

- [ ] Transfer repo to `McTalian-WoW-Addons`
- [ ] Update local git remote
- [ ] Delete repo-level secrets now covered by org-level (`GH_PAT`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`)
- [ ] Re-add branch protection bypass as org owner/admin (bypass lists don't transfer)
- [ ] Update all `McTalian/` references to `McTalian-WoW-Addons/` in workflows, TOC, README, package.json, etc.
- [ ] Replace inline workflows with thin callers to WBT reusable workflows
- [ ] Add `dependabot.yml` with `github-actions` ecosystem (and `uv` if Python tooling present)
- [ ] Verify CI passes with org-level secrets

**After all addons are migrated:**

- [ ] Retire old personal PAT (`McTalian` account fine-grained token)

---

## WBT: `wow-build-tools init` Command

Scaffolding subcommand for onboarding new addons quickly.

```bash
wow-build-tools init --name MyAddon --flavors retail,classic --platforms curseforge,wago,wowi
```

**Would generate:**

- Thin caller workflows (`.github/workflows/`) pointing to WBT reusable workflows
- `dependabot.yml` with `github-actions` ecosystem
- `Makefile` (parameterized)
- Rockspec
- `.luacov`
- Basic addon skeleton (`MyAddon/MyAddon.toc`, `MyAddon/Core.lua`)

- [ ] Design the command interface and template system
- [ ] Implement template files
- [ ] Add to wow-build-tools CLI

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

Run `pinact` as a one-shot local operation to replace mutable tag refs (e.g. `@v4`, `@v1-beta`) with pinned commit SHAs in all workflow files. Dependabot (`github-actions` ecosystem, now configured on all repos) will keep them updated from there.

**Repos to run against:** `Endeavoring`, `RPGLootFeed`, `wow-build-tools`

- [ ] Run `pinact run` across all three repos
- [ ] Commit and push the pinned refs

---

## Low Priority: Minor Config Alignment

Small inconsistencies between addon repos, not worth a dedicated sprint but worth cleaning up when touching these files.

| File                    | Status                                         |
| ----------------------- | ---------------------------------------------- |
| `.luacov`               | Minor differences between repos — low priority |
| Rockspec                | Minor differences — low priority               |
| Makefile (core targets) | Minor differences — low priority               |
