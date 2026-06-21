<!-- filepath: /home/mctalian/code/wow-build-tools/toc/update/README.md -->

# WoW Build Tools - TOC Update

GitHub Action to automatically update the interface version(s) in your World of Warcraft addon TOC files.

## Usage

```yaml
- name: Update TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
```

## What It Does

This action updates your addon TOC files with the latest interface versions based on the selected release channel(s). It automatically detects all TOC files in your addon directory and updates the appropriate interface version fields.

## Inputs

### `args`

**Optional** - Additional arguments to pass to `wow-build-tools toc update`.

**Default**: `""`

## Available Arguments

- `--ptr` / `-p`: Update to PTR interface versions
- `--beta` / `-b`: Update to Beta interface versions

## Examples

### Basic usage

```yaml
- name: Update TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
```

### Update for PTR versions

```yaml
- name: Update TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
  with:
    args: "--ptr"
```

### Update for Beta versions

```yaml
- name: Update TOC files
  uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1
  with:
    args: "--beta"
```

### Automatic updates with pull request

```yaml
name: Update TOC

on:
  schedule:
    - cron: 0 12 * * * # Runs daily at 12:00 UTC
  workflow_dispatch:

jobs:
  update-toc:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Update TOC files
        uses: McTalian-WoW-Addons/wow-build-tools/toc/update@v1

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          commit-message: "chore: update TOC interface versions"
          title: "Update TOC interface versions"
          branch: update-toc
```

## License

MIT
