# Deprecation Tracker

Tracking features and APIs that are deprecated and planned for removal in v2.

## Deprecated Features

### `toc/update` GitHub Action - `args` Input

**Deprecated:** v1.x  
**Planned Removal:** v2.0  
**Reason:** Replaced with individual, type-safe inputs for better usability and validation  
**Migration Path:** Replace `args` input with individual inputs:

| Old Format            | New Inputs              |
| --------------------- | ----------------------- |
| `args: -a addon-name` | `addon-dir: addon-name` |
| `args: -b`            | `beta: true`            |
| `args: -p`            | `ptr: true`             |
| `args: -v`            | `debug: true`           |
| `args: -V`            | `verbose: true`         |
| `args: --no-color`    | `no-color: true`        |
| `args: --no-emoji`    | `no-emoji: true`        |

**Current Behavior:** Using `args` input will emit a GitHub Actions deprecation warning.

**Example Migration:**

Before:

```yaml
- uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
  with:
    args: -a MyAddon -b -p
```

After:

```yaml
- uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
  with:
    addon-dir: MyAddon
    beta: true
    ptr: true
```
