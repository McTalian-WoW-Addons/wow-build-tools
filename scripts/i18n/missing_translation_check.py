r"""
missing_translation_check.py

Compares each non-enUS locale file against enUS.lua and outputs per-locale
markdown reports listing missing translation keys. Reports are written to
--output-dir for consumption by create_or_update_i18n_issues.py.

Exits non-zero if any locale file contains keys not present in enUS.lua
(i.e. extra keys that should be removed).

Usage:
    uv run --with defusedxml --no-project missing_translation_check.py \\
        --locale-dir MyAddon/locale \\
        [--locale-xml index.xml] \\
        [--output-dir .i18n-output] \\
        [--ignored-files main.lua init.lua load.lua]
"""

import argparse
import os
import re
import sys
import textwrap

import defusedxml.ElementTree as ET


def parse_args():
    parser = argparse.ArgumentParser(
        description=(
            "Compare each locale against enUS and output per-locale "
            "missing-key markdown reports."
        )
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
    parser.add_argument(
        "--output-dir",
        default=".i18n-output",
        help="Directory for output markdown report files (default: .i18n-output)",
    )
    parser.add_argument(
        "--ignored-files",
        nargs="*",
        default=["main.lua", "init.lua", "load.lua"],
        help="Locale filenames to skip (default: main.lua init.lua load.lua)",
    )
    return parser.parse_args()


def parse_locales_xml(xml_file):
    """Parse the locale XML file and return the list of Lua filenames."""
    tree = ET.parse(xml_file)
    root = tree.getroot()
    namespace = {"ns": root.tag.split("}")[0].strip("{")}
    return [script.attrib["file"] for script in root.findall("ns:Script", namespace)]


def load_lua_file(lua_file):
    """
    Load a Lua locale file into a {key: value} dict.
    Handles both double-quoted and single-quoted (with escape sequences) string values,
    as well as the bare `true` marker used for the reference locale.
    """
    result = {}
    with open(lua_file, "r") as file:
        for line in file:
            # Match: L["key"] = true | "value" | 'value (with escapes)'
            match = re.match(
                r"""L\["([^"]+)"\]\s*=\s*(true|"[^"]*"|'(?:[^'\\]|\\.)*')""",
                line.strip(),
            )
            if match:
                result[match[1]] = match[2]
    return result


def compare_translations(reference_dict, target_dict, locale, locale_dir, repo):
    """
    Compare target locale dict against the reference (enUS) dict.
    Returns (markdown_report_or_None, list_of_extra_keys).
    """
    missing_keys = []
    extra_keys = []

    for key, value in reference_dict.items():
        if key not in target_dict:
            enUS_value = key if value.lower() == "true" else value.strip("\"'")
            missing_keys.append(f"| {key} | {enUS_value} |")

    for key in target_dict:
        if key not in reference_dict:
            extra_keys.append(key)

    if missing_keys:
        markdown_report = f"# Translation Status for {locale}\n\n"
        markdown_report += (
            f"Translation progress: "
            f"{(1 - (len(missing_keys) / len(reference_dict))) * 100:.1f}%\n\n"
        )
        markdown_report += f"Missing translations: {len(missing_keys)}\n\n"
        markdown_report += "<details>\n"
        markdown_report += (
            "    <summary>Missing Keys and their enUS values</summary>\n\n"
        )
        markdown_report += "| Missing Key | enUS Value |\n"
        markdown_report += "|-------------|------------|\n"
        markdown_report += "\n".join(missing_keys)
        markdown_report += "\n</details>\n\n"

        if repo:
            markdown_report += (
                f"\n\n_You can even make changes for "
                f"[this file](https://github.com/{repo}/edit/main/{locale_dir}/{locale})"
                f" and open a PR directly in your browser_\n\n"
            )

        translation_stub = "\n".join(
            [f'L["{key.split("|")[1].strip()}"] = ""' for key in missing_keys]
        )
        details_section = textwrap.dedent(f"""

<details>
    <summary>Please provide one or more of these values in a Pull Request or a Comment on this issue</summary>

```
{translation_stub}
```
</details>

""")
        markdown_report += details_section
    else:
        markdown_report = None

    return markdown_report, extra_keys


def main():
    args = parse_args()
    locale_dir = args.locale_dir
    output_dir = args.output_dir
    ignored_files = args.ignored_files
    locales_xml = os.path.join(locale_dir, args.locale_xml)
    repo = os.environ.get("GITHUB_REPOSITORY")

    print("=" * 60)
    print("missing_translation_check.py")
    print("=" * 60)
    print(f"  CWD:         {os.path.abspath(os.curdir)}")
    print(f"  locale-dir:  {os.path.abspath(locale_dir)}")
    print(f"  locale-xml:  {os.path.abspath(locales_xml)}")
    print(f"  output-dir:  {os.path.abspath(output_dir)}")
    print(f"  repo:        {repo or '(not set)'}")
    print(f"  ignored:     {ignored_files}")
    print()

    # Recreate output directory (clear previous reports)
    os.makedirs(output_dir, exist_ok=True)
    for filename in os.listdir(output_dir):
        file_path = os.path.join(output_dir, filename)
        try:
            os.unlink(file_path)
        except Exception as e:
            print(f"Failed to delete {file_path}. Reason: {e}")

    locale_files = parse_locales_xml(locales_xml)
    print(f"Locale files found in {args.locale_xml}: {locale_files}")

    reference_file = "enUS.lua"
    reference_path = os.path.join(locale_dir, reference_file)
    reference_dict = load_lua_file(reference_path)
    print(
        f"Reference (enUS): {os.path.abspath(reference_path)} — {len(reference_dict)} keys"
    )
    print()

    has_extra_keys = False

    for locale_file in locale_files:
        if locale_file in ignored_files or locale_file == reference_file:
            continue

        locale_path = os.path.join(locale_dir, locale_file)
        print(f"Checking {os.path.abspath(locale_path)} ...", end=" ")
        target_dict = load_lua_file(locale_path)
        print(f"{len(target_dict)} keys")
        markdown_report, extra_keys = compare_translations(
            reference_dict, target_dict, locale_file, locale_dir, repo
        )

        if markdown_report:
            output_file_path = os.path.join(
                output_dir, f"{locale_file}_missing_keys.md"
            )
            with open(output_file_path, "w") as output_file:
                output_file.write(markdown_report)
            print(f"  Missing translations written to {output_file_path}")

        if extra_keys:
            print(f"\n\nERROR: Extra translation keys in {locale_file}:")
            for key in extra_keys:
                print(f"  {key}")
            has_extra_keys = True

    print()
    if has_extra_keys:
        sys.exit(1)
    else:
        print("No extra translation keys found.")


if __name__ == "__main__":
    main()
