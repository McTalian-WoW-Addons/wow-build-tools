"""
hardcode_string_check.py

Scans .lua files for hard-coded strings that should be localized:
  - Print(...) calls with literal string arguments
  - Ace3-style config entries with hard-coded `name` or `desc` fields
  - Ace3-style `values` tables with hard-coded key-value pairs

Exits non-zero if any hard-coded strings are found.

Usage:
    uv run --no-project hardcode_string_check.py \\
        --addon-dir MyAddon \\
        [--ignore-files file1.lua file2.lua] \\
        [--ignore-dirs icons locale libs]
"""

import argparse
import os
import re


def parse_args():
    parser = argparse.ArgumentParser(
        description="Scan .lua files for hard-coded strings that should be localized.",
    )
    parser.add_argument(
        "--addon-dir",
        required=True,
        help="Path to the addon directory to scan (e.g., MyAddon)",
    )
    parser.add_argument(
        "--ignore-files",
        nargs="*",
        default=[],
        help="Lua filenames to ignore",
    )
    parser.add_argument(
        "--ignore-dirs",
        nargs="*",
        default=["icons", "locale"],
        help="Directory names to ignore (default: icons locale)",
    )
    return parser.parse_args()


def should_ignore(path, ignore_files, ignore_dirs):
    if os.path.basename(path) in ignore_dirs:
        return True
    for d in ignore_dirs:
        if path.startswith(f"./{d}/"):
            return True
    return os.path.basename(path) in ignore_files


def check_hardcoded_strings(file_content, filename):
    """Return a list of hard-coded string issues found in file_content."""
    issues = []

    # Strip lines with inline suppression comment before any checks.
    # Use `-- nocheck` on a line to suppress false positives.
    lines = file_content.splitlines()
    file_content = "\n".join(line for line in lines if "-- nocheck" not in line)

    # Print(...) calls with hard-coded string arguments
    print_matches = re.findall(
        r'(?:\w+:)?Print\(\s*"([^"]+)(?:"(?:\s*\+|\s*\.\.)\s*|\s*"\s*\))',
        file_content,
        re.DOTALL,
    )
    for match in print_matches:
        issues.append(
            f'Hard-coded string in Print(...) in {filename}: "{match}"'
        )

    # Ace3 config entries with hard-coded name/desc fields
    config_matches = re.findall(
        r'\b(name|desc)\s*=\s*"([^"]+)"', file_content, re.DOTALL
    )
    for field, value in config_matches:
        issues.append(f'Hard-coded {field} in {filename}: "{value}"')

    # Ace3 config `values` tables with hard-coded string entries
    values_matches = re.findall(r"\bvalues\s*=\s*{([^}]*)}", file_content, re.DOTALL)
    for match in values_matches:
        key_value_matches = re.findall(r'\[?"?(.*)"?\]?\s*=\s*"([^"]+)"', match)
        for key, value in key_value_matches:
            issues.append(
                f'Hard-coded key-value pair in "values" table in {filename}: '
                f'"{key.strip()} = {value.strip()}"'
            )

    return issues


def scan_directory(directory, ignore_files=None, ignore_dirs=None):
    """Recursively scan directory for .lua files and check for hard-coded strings."""
    if ignore_files is None:
        ignore_files = []
    if ignore_dirs is None:
        ignore_dirs = []
    all_issues = []

    for root, dirs, files in os.walk(directory):
        # Prune ignored directories in-place so os.walk won't descend into them
        dirs[:] = [
            d
            for d in dirs
            if not should_ignore(os.path.join(root, d), [], ignore_dirs)
        ]

        for file in files:
            if file.endswith(".lua") and not should_ignore(
                os.path.join(root, file), ignore_files, []
            ):
                filepath = os.path.join(root, file)
                with open(filepath, "r", encoding="utf-8") as f:
                    content = f.read()
                issues = check_hardcoded_strings(content, filepath)
                if issues:
                    all_issues.extend(issues)

    return all_issues


if __name__ == "__main__":
    args = parse_args()

    print("=" * 60)
    print("hardcode_string_check.py")
    print("=" * 60)
    print(f"  CWD:          {os.path.abspath(os.curdir)}")
    print(f"  addon-dir:    {os.path.abspath(args.addon_dir)}")
    print(f"  ignore-files: {args.ignore_files or '(none)'}")
    print(f"  ignore-dirs:  {args.ignore_dirs}")
    print()

    issues = scan_directory(args.addon_dir, args.ignore_files, args.ignore_dirs)

    if issues:
        print("Hard-coded strings found:")
        for issue in issues:
            print(f"  {issue}")
        exit(1)
    else:
        print("No hard-coded strings found.")
