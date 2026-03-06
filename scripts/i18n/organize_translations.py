"""
organize_translations.py

Sorts locale file entries within --#region / --#endregion blocks:
  1. Untranslated (commented-out) entries first, sorted alphabetically
  2. Translated entries next, sorted alphabetically

Also ensures each non-enUS locale file has all regions from enUS.lua,
adding commented-out stubs for any missing keys.

Modifies locale files in-place. Safe to run repeatedly (idempotent).

Usage:
    uv run --with defusedxml --no-project organize_translations.py \\
        --locale-dir MyAddon/locale \\
        [--locale-xml index.xml]
"""

import argparse
import os
import re
from collections import OrderedDict

import defusedxml.ElementTree as ET


def parse_args():
    parser = argparse.ArgumentParser(
        description="Sort locale file entries within --#region blocks alphabetically.",
    )
    parser.add_argument(
        "--locale-dir",
        required=True,
        help="Path to the locale directory (e.g., MyAddon/locale)",
    )
    parser.add_argument(
        "--locale-xml",
        default="index.xml",
        help="XML file listing locale scripts within locale-dir (default: index.xml)",
    )
    return parser.parse_args()


def parse_locales_xml(xml_file):
    """Parse the locale XML file and return the list of Lua filenames."""
    tree = ET.parse(xml_file)
    root = tree.getroot()
    namespace = {"ns": root.tag.split("}")[0].strip("{")}
    return [
        script.attrib["file"] for script in root.findall("ns:Script", namespace)
    ]


def parse_locale_file_with_regions(file_path):
    """
    Parse a Lua locale file into header, regions, and footer sections.

    Returns:
        header_lines: Lines before the first --#region
        regions: OrderedDict of region_name -> list of (key, value, original_line)
        footer_lines: Lines after the last --#endregion
    """
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # Extract header (everything before the first --#region)
    header_match = re.search(r"(.*?)^--#region", content, re.DOTALL | re.MULTILINE)
    header_lines = (
        header_match.group(1).splitlines() if header_match else content.splitlines()
    )

    # Extract all region blocks
    region_pattern = r"^--#region\s+([^\n]+)\n(.*?)--#endregion"
    region_matches = re.findall(region_pattern, content, re.DOTALL | re.MULTILINE)

    regions = OrderedDict()
    for region_name, region_content in region_matches:
        translations = []
        for line in region_content.splitlines():
            # Skip blank lines and pure comment lines (but keep --[[ ]])
            if not line.strip() or (
                line.strip().startswith("--") and "--[[" not in line
            ):
                continue

            match = re.match(r'L\["(.+?)"\]\s*=\s*(.+)', line.strip())
            if match:
                key = match.group(1)
                value = match.group(2)
                translations.append((key, value, line))

        regions[region_name.strip()] = translations

    # Extract footer (everything after the last --#endregion)
    parts = content.split("--#endregion\n")
    if len(parts) > 1:
        footer_lines = parts[-1].splitlines()
        # Strip leading blank lines to prevent accumulation on repeated runs
        while footer_lines and not footer_lines[0].strip():
            footer_lines.pop(0)
    else:
        footer_lines = []

    return header_lines, regions, footer_lines


def sort_regions_by_keys(regions):
    """
    Sort entries within each region:
      1. Commented-out (untranslated) entries first, alphabetically by key
      2. Active (translated) entries next, alphabetically by key
    """
    for region_name, translations in regions.items():
        untranslated = []
        translated = []

        for entry in translations:
            _key, _value, line = entry
            if line.strip().startswith("--"):
                untranslated.append(entry)
            else:
                translated.append(entry)

        untranslated.sort(key=lambda x: x[0].lower())
        translated.sort(key=lambda x: x[0].lower())

        regions[region_name] = untranslated + translated

    return regions


def generate_updated_locale_file(header_lines, regions, footer_lines):
    """Reconstruct the locale file content from its parsed components."""
    lines = header_lines.copy()

    for region_name, translations in regions.items():
        lines.append(f"--#region {region_name}")
        for _key, _value, original_line in translations:
            lines.append(original_line)
        lines.append("--#endregion")
        lines.append("")  # Blank line after each endregion

    lines.extend(footer_lines)

    # Remove trailing blank lines, ensure exactly one trailing newline
    while lines and not lines[-1].strip():
        lines.pop()

    return "\n".join(lines) + "\n"


def process_locale_file(reference_regions, locale_file_path):
    """
    Update a locale file to:
      1. Add any region blocks missing from the reference (enUS)
      2. Add commented-out stubs for keys missing in this locale
      3. Sort entries alphabetically within each region
    """
    if not os.path.exists(locale_file_path):
        print(f"File not found: {locale_file_path}")
        return

    header_lines, target_regions, footer_lines = parse_locale_file_with_regions(
        locale_file_path
    )

    # Build a flat lookup of existing translations in this locale
    existing_translations = {}
    for region_name, translations in target_regions.items():
        for key, value, _ in translations:
            existing_translations[key] = (value, region_name)

    # Rebuild regions to match the reference structure
    updated_regions = OrderedDict()

    for ref_region_name, ref_translations in reference_regions.items():
        region_translations = []

        for ref_key, ref_value, _ in ref_translations:
            if ref_key in existing_translations:
                target_value, _target_region = existing_translations[ref_key]
                region_translations.append(
                    (ref_key, target_value, f'L["{ref_key}"] = {target_value}')
                )
            else:
                # Key not yet translated — add a commented-out stub with the English value
                region_translations.append(
                    (ref_key, "", f'-- L["{ref_key}"] = {ref_value}')
                )

        updated_regions[ref_region_name] = region_translations

    updated_regions = sort_regions_by_keys(updated_regions)
    updated_content = generate_updated_locale_file(
        header_lines, updated_regions, footer_lines
    )

    with open(locale_file_path, "w", encoding="utf-8") as f:
        f.write(updated_content)


def main():
    args = parse_args()
    locale_dir = args.locale_dir
    locales_xml = os.path.join(locale_dir, args.locale_xml)

    print("=" * 60)
    print("organize_translations.py")
    print("=" * 60)
    print(f"  CWD:        {os.path.abspath(os.curdir)}")
    print(f"  locale-dir: {os.path.abspath(locale_dir)}")
    print(f"  locale-xml: {os.path.abspath(locales_xml)}")
    print()

    locale_files = parse_locales_xml(locales_xml)
    print(f"Locale files found in {args.locale_xml}: {locale_files}")

    reference_file = "enUS.lua"
    reference_path = os.path.join(locale_dir, reference_file)
    ignored_files = ["main.lua", "init.lua", "load.lua"]

    print(f"Sorting reference file: {os.path.abspath(reference_path)}")

    # Parse and sort the reference file (enUS.lua) first
    header_lines, reference_regions, footer_lines = parse_locale_file_with_regions(
        reference_path
    )
    sorted_reference_regions = sort_regions_by_keys(reference_regions)

    updated_reference_content = generate_updated_locale_file(
        header_lines, sorted_reference_regions, footer_lines
    )
    with open(reference_path, "w", encoding="utf-8") as f:
        f.write(updated_reference_content)

    # Process each non-reference locale file
    for locale_file in locale_files:
        if locale_file in ignored_files or locale_file == reference_file:
            continue

        locale_path = os.path.join(locale_dir, locale_file)
        print(f"Processing: {os.path.abspath(locale_path)}")
        process_locale_file(sorted_reference_regions, locale_path)

    print("Done.")


if __name__ == "__main__":
    main()
