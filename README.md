# wow-build-tools

<!-- markdownlint-disable-next-line MD033 -->
<img src="icon.jpg" alt="wow-build-tools" width="200">

This repository aims to be a collection of tools to help with the development of World of Warcraft addons. The main focus is on speed and ease of use, with the goal of making the development process as smooth as possible.

## Installation

Once this project is out of beta, the installation process should be a one-time thing as the tool is written to self update on non-pre-release versions. For now, you'll need to manually download the latest version of the tool when new versions are released.

### Local

Head over to the [releases page](https://github.com/McTalian-WoW-Addons/wow-build-tools/releases) and download the latest release for your operating system and architecture.

- **Windows?** it's likely going to be the `wow-build-tools_windows_amd64.zip` file.
- **Linux or WSL?** you'll likely want the `wow-build-tools_linux_amd64.zip` file.
- **macOS?** you'll want to double-check if you have an Intel (`darwin_amd64.zip`) or Apple Silicon (`darwin_arm64.zip`) processor, and download the appropriate file.

Extract the contents of the zip file to a directory of your choosing, and add that directory to your PATH if you'd like to use the tool from anywhere on your system.

## GitHub Actions

If you're looking to use `wow-build-tools` in a GitHub Action, you can use the following example workflow:

```yaml
name: Package and release

on:
  release:
    types:
      - published

permissions: {}

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    env:
      CF_API_KEY: ${{ secrets.CF_API_KEY }} # CurseForge API Key, required for uploading to CurseForge
      WOWI_API_TOKEN: ${{ secrets.WOWI_API_TOKEN }} # WoWInterface API Token, required for uploading to WoWInterface
      WAGO_API_TOKEN: ${{ secrets.WAGO_API_TOKEN }} # Wago.io API Token, required for uploading to Wago.io

    steps:
      - name: Clone project
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Package and release
        env:
          # GitHub Token, required for creating releases
          GITHUB_OAUTH: ${{ secrets.GITHUB_TOKEN }}
        uses: McTalian-WoW-Addons/wow-build-tools@v1-beta
```

## Features

I have many plans for this project, and I will communicate those plans at a later date.

### `BigWigsMods/packager` feature parity

To start, I'd like `wow-build-tools` to be as close to a drop-in replacement for `BigWigsMods/packager` as possible. That means, ideally, it should be able to handle all of the same features as `packager` with the same level of ease and speed or better.

For feature parity progress, please see the [BigWigsMods/packager feature parity wiki page](https://github.com/McTalian-WoW-Addons/wow-build-tools/wiki/BigWigs-Packager-Parity)

### Additional features

In addition to feature parity with `BigWigsMods/packager`, I have a few ideas for additional features that I think would be useful for addon authors:

- [x] Autoupdating the tool itself
- [ ] More token replacements
- [x] Use GitHub Release contents as a source for the changelog
- [ ] Guided tour of the tool
- [ ] Various warnings and checks to help catch issues with the addon before packaging
  - [x] Missing embedded library Curse attributions
- [ ] Monorepo support
- [x] Automatic propagation of addon changes to all installed and compatible game versions (available via `config` and `link` commands)
- [ ] Option to create a Lua version of the changelog
- [ ] New Addon Scaffolding
- [ ] A badge to proudly display that your addon is built with `wow-build-tools`!

## Link Command

The `wow-build-tools build link` command creates symlinks from your built addon(s) into configured WoW client AddOns directories.

By default, linking is TOC-interface aware and only targets compatible client installs.

Compatibility lookup order:

1. TOC files in the project top directory
2. TOC files in the release directory
3. TOC files inside each addon folder in the release directory

If no TOC compatibility can be derived, the command falls back to linking all selected installs.

### Common usage

```bash
# Link using local .release output and TOC-compatible client filtering
wow-build-tools build link

# Force linking into all selected clients (skip TOC compatibility filtering)
wow-build-tools build link --all-flavors

# WSL workflow: use the Windows path to your release directory
wow-build-tools.exe build link -w "C:\\path\\to\\addon\\.release"

# Restrict linking to specific configured clients
wow-build-tools build link --flavor retail --flavor classicEra
```

If a destination already exists as a symlink, it is replaced automatically.
Use `--force` only when overwriting an existing non-symlink file or directory.

## Watch Command

The `wow-build-tools build watch` command rebuilds when addon files change.

Because `watch` is a subcommand of `build`, it also supports build flags like forced build types:

```bash
# Typical active development loop: always build in alpha mode
wow-build-tools build watch --force-alpha

# Alternative forced modes
wow-build-tools build watch --force-beta
wow-build-tools build watch --force-dev
```

## Inspiration and acknowledgements

My main inspiration comes from my desire to always make developer experience as smooth as possible. I've had a few different roles across different companies and industries that have focused on developer experience, and I've always found it to be a rewarding challenge and a force multiplier for teams. What better way to give back to the WoW community than to align my passions and expertise to help make the development process easier for addon authors?

I was also heavily inspired by [BigWigsMods/packager](https://github.com/BigWigsMods/packager) which provides (to my knowledge) the most widely used tool for packaging addons for distribution via CurseForge, WoWInterface, and Wago.io. Thank you to the authors and contributors of that project for all of their hard work and dedication!
