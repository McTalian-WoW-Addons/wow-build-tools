# WoW Build Tools - TOC Check

GitHub Action to check for common issues in your World of Warcraft addon TOC files.

## Usage

```yaml
- name: Check TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/check@v1
```

## What It Checks

This action validates your addon TOC files for:

- **Name consistency**: Verifies the TOC file name matches the addon folder name
- **Missing files**: Identifies `.lua` files present on disk but not included in TOC or XML include trees
- **Interface versions**: Detects outdated interface versions and warns when updates are available

## Inputs

### `args`

**Optional** - Additional arguments to pass to `wow-build-tools toc check`.

**Default**: `""`

## Available Arguments

- `--ignore` / `-x`: Files to ignore during the check (if XML files are provided, their specified includes will also be ignored)
- `--skip-interface-check`: Skip checking the interface version
- `--skip-missing-files-check`: Skip checking for missing files
- `--skip-name-check`: Skip checking that the TOC file and addon folder have the same name
- `--ptr` / `-p`: Check against PTR interface versions
- `--beta` / `-b`: Check against Beta interface versions

## Examples

### Basic usage

```yaml
- name: Check TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/check@v1
```

### Ignore specific files

```yaml
- name: Check TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/check@v1
  with:
    args: "--ignore test.lua --ignore docs.lua"
```

### Skip specific checks

```yaml
- name: Check TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/check@v1
  with:
    args: "--skip-interface-check"
```

### Check against PTR versions

```yaml
- name: Check TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/check@v1
  with:
    args: "--ptr"
```

## License

MIT
