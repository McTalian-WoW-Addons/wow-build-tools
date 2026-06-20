r"""
create_or_update_i18n_issues.py

Reads per-locale markdown reports from --output-dir (produced by
missing_translation_check.py) and creates or updates GitHub issues
for each locale with an "i18n" label.

Repo is read from the GITHUB_REPOSITORY environment variable
(automatically set by GitHub Actions).
GITHUB_TOKEN must be set for API authentication.

Usage:
    uv run --with requests --no-project create_or_update_i18n_issues.py \\
        [--output-dir .i18n-output]
"""

import argparse
import os

import requests

GITHUB_API_URL = "https://api.github.com"
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")

headers = {
    "Authorization": f"Bearer {GITHUB_TOKEN}",
    "Accept": "application/vnd.github.v3+json",
}

DEFAULT_GET_TIMEOUT = 10
DEFAULT_PATCH_TIMEOUT = 10


def parse_args():
    parser = argparse.ArgumentParser(
        description=(
            "Create or update GitHub issues for each locale with missing "
            "translation reports."
        )
    )
    parser.add_argument(
        "--output-dir",
        default=".i18n-output",
        help="Directory containing markdown report files (default: .i18n-output)",
    )
    return parser.parse_args()


def get_repo():
    """Read the GitHub repository (owner/name) from the GITHUB_REPOSITORY env var."""
    github_repository = os.environ.get("GITHUB_REPOSITORY")
    if not github_repository:
        raise ValueError("GITHUB_REPOSITORY environment variable is not set")
    parts = github_repository.split("/")
    if len(parts) != 2:
        raise ValueError(f"Invalid GITHUB_REPOSITORY format: {github_repository}")
    return parts[0], parts[1]


def get_all_translation_issues(repo_owner, repo_name):
    """Search for open issues with the 'i18n' label and return a dict keyed by locale."""
    search_url = f"{GITHUB_API_URL}/search/issues"
    query = f"repo:{repo_owner}/{repo_name} is:issue label:i18n state:open"
    params = {"q": query}

    response = requests.get(
        search_url, headers=headers, params=params, timeout=DEFAULT_GET_TIMEOUT
    )
    response.raise_for_status()
    issues = response.json().get("items", [])

    issues_dict = {}
    for issue in issues:
        title = issue["title"]
        # Strip the 'i18n: ' prefix and ' Translations' suffix to extract the locale
        if title.startswith("i18n: ") and title.endswith(" Translations"):
            locale = title[len("i18n: ") : -len(" Translations")]
            issues_dict[locale] = issue

    return issues_dict


def create_issue(repo_owner, repo_name, locale, markdown_content):
    """Create a new GitHub issue for the given locale."""
    issue_url = f"{GITHUB_API_URL}/repos/{repo_owner}/{repo_name}/issues"
    issue_data = {
        "title": f"i18n: {locale} Translations",
        "body": markdown_content,
        "labels": ["i18n"],
    }

    response = requests.post(
        issue_url, headers=headers, json=issue_data, timeout=DEFAULT_GET_TIMEOUT
    )
    response.raise_for_status()
    print(f"Issue created: {response.json().get('html_url')}")


def update_issue(repo_owner, repo_name, issue_number, markdown_content):
    """Update an existing GitHub issue with new markdown content."""
    issue_url = f"{GITHUB_API_URL}/repos/{repo_owner}/{repo_name}/issues/{issue_number}"

    response = requests.patch(
        issue_url,
        headers=headers,
        json={"body": markdown_content},
        timeout=DEFAULT_PATCH_TIMEOUT,
    )
    response.raise_for_status()
    print(f"Issue updated: {response.json().get('html_url')}")


def process_markdown_files(output_directory, repo_owner, repo_name):
    """Process each markdown report and create or update the corresponding GitHub issue."""
    issues_dict = get_all_translation_issues(repo_owner, repo_name)

    for filename in os.listdir(output_directory):
        if filename.endswith("_missing_keys.md"):
            locale = filename.split(".")[0]
            with open(os.path.join(output_directory, filename), "r") as file:
                markdown_content = file.read()

            existing_issue = issues_dict.get(locale)
            if existing_issue:
                update_issue(
                    repo_owner, repo_name, existing_issue["number"], markdown_content
                )
            else:
                create_issue(repo_owner, repo_name, locale, markdown_content)


if __name__ == "__main__":
    args = parse_args()
    repo_owner, repo_name = get_repo()

    print("=" * 60)
    print("create_or_update_i18n_issues.py")
    print("=" * 60)
    print(f"  CWD:        {os.path.abspath(os.curdir)}")
    print(f"  output-dir: {os.path.abspath(args.output_dir)}")
    print(f"  repo:       {repo_owner}/{repo_name}")
    print()

    process_markdown_files(args.output_dir, repo_owner, repo_name)
