r"""
check_for_missing_locale_keys.py

Verifies that every locale key used in addon Lua source code has a
corresponding definition in enUS.lua. Also reports unused keys defined
in enUS.lua.

Matches any access to the locale table regardless of namespace prefix:
  L["key"]         (bare)
  ns.L["key"]      (BUF / Endeavoring style)
  G_RLF.L["key"]   (RPGLootFeed style)

Exits non-zero if any keys are used in code but not defined in enUS.lua.

Usage:
    uv run --no-project check_for_missing_locale_keys.py \\
        --addon-dir MyAddon \\
        --locale-dir MyAddon/locale \\
        [--locale-table L]
"""

import argparse
import os
import re
from collections import Counter


def parse_args():
    parser = argparse.ArgumentParser(
        description=(
            "Verify all locale key call sites in addon code have a matching "
            "definition in enUS.lua."
        )
    )
    parser.add_argument(
        "--addon-dir",
        required=True,
        help="Path to the addon directory to scan for locale key usage (e.g., MyAddon)",
    )
    parser.add_argument(
        "--locale-dir",
        required=True,
        help="Path to the locale directory containing enUS.lua (e.g., MyAddon/locale)",
    )
    parser.add_argument(
        "--locale-table",
        default="L",
        help=(
            "Name of the locale table (default: L). "
            'Matches bare usage (L["key"]) and any namespace prefix '
            '(ns.L["key"], G_RLF.L["key"], etc.) automatically.'
        ),
    )
    return parser.parse_args()


comment_pattern = re.compile(r"^\s*--")
# Matches the definition side: L["key"] = ...
definition_pattern = re.compile(r'L\["(.*?)"\]')


def get_locale_keys(addon_dir, locale_dir, table_name):
    """
    Scan all .lua files in addon_dir (excluding locale_dir) for locale key usage.
    Matches bare usage (L["key"]) and any namespace prefix (ns.L["key"],
    G_RLF.L["key"], etc.) by anchoring on a word boundary before the table name.
    """
    locale_key_pattern = re.compile(rf'\b{re.escape(table_name)}\["(.*?)"\]')
    locale_keys = set()
    abs_locale_dir = os.path.abspath(locale_dir)
    files_scanned = 0

    for root, _, files in os.walk(addon_dir):
        # Skip the locale directory itself — we only scan usage sites here
        if abs_locale_dir in os.path.abspath(root):
            continue
        for file in files:
            if file.endswith(".lua"):
                filepath = os.path.join(root, file)
                files_scanned += 1
                with open(filepath, "r") as f:
                    for line in f:
                        if not comment_pattern.match(line):
                            locale_keys.update(locale_key_pattern.findall(line))

    print(
        f"  Scanned {files_scanned} .lua source file(s), found {len(locale_keys)} unique key usage(s)"
    )
    return locale_keys


def get_defined_keys(enUS_file):
    """
    Read all L["key"] definitions from enUS.lua.
    Reports duplicate definitions as a warning.
    """
    defined_keys = set()
    raw_keys = []

    with open(enUS_file, "r") as f:
        for line in f:
            if not comment_pattern.match(line):
                keys = definition_pattern.findall(line)
                defined_keys.update(keys)
                raw_keys.extend(keys)

    key_counts = Counter(raw_keys)
    duplicates = [key for key, count in key_counts.items() if count > 1]
    if duplicates:
        print("Duplicate keys defined in enUS.lua:")
        for key in duplicates:
            print(f'  L["{key}"]')
        print("\nPlease remove the duplicate keys from enUS.lua")

    return defined_keys


def check_missing_keys(addon_dir, locale_dir, table_name):
    enUS_file = os.path.join(locale_dir, "enUS.lua")

    print("=" * 60)
    print("check_for_missing_locale_keys.py")
    print("=" * 60)
    print(f"  CWD:          {os.path.abspath(os.curdir)}")
    print(f"  addon-dir:    {os.path.abspath(addon_dir)}")
    print(f"  locale-dir:   {os.path.abspath(locale_dir)}")
    print(f"  enUS file:    {os.path.abspath(enUS_file)}")
    print(
        f'  locale-table: {table_name}  (matches: {table_name}["key"], ns.{table_name}["key"], etc.)'
    )
    print()

    locale_keys = get_locale_keys(addon_dir, locale_dir, table_name)
    defined_keys = get_defined_keys(enUS_file)
    print(f"  Defined keys in enUS.lua: {len(defined_keys)}")

    missing_keys = locale_keys - defined_keys
    unused_keys = defined_keys - locale_keys

    if unused_keys:
        print("Possibly unused locale keys defined in enUS.lua:\n")
        for key in unused_keys:
            print(f'  L["{key}"]')
        print("\nPlease remove the extra keys from enUS.lua\n")

    if missing_keys:
        print("Missing locale keys in enUS.lua:\n")
        for key in missing_keys:
            print(f'  L["{key}"]')
        print("\nPlease define the missing keys in enUS.lua")
        exit(1)
    else:
        print("All locale keys are defined in enUS.lua")
        exit(0)


if __name__ == "__main__":
    args = parse_args()
    check_missing_keys(args.addon_dir, args.locale_dir, args.locale_table)
