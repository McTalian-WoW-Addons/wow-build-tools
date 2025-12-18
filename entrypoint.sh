#!/bin/bash
set -e

# GitHub Actions sets the workspace to /github/workspace
cd /github/workspace

# Ensure we're running in the GitHub Actions environment
export CI="true"
export GITHUB_ACTIONS="true"

# Get arguments passed from action.yml args section
# $1 contains the args input from the action
# If first arg starts with /, it's a subcommand (e.g., /toc/check)
if [[ "$1" == /* ]]; then
  # Remove leading slash and convert to command
  # /toc/check -> toc check
  COMMAND="${1#/}"
  COMMAND="${COMMAND//\// }"
  shift
  echo "Running: wow-build-tools $COMMAND ${@}"
  wow-build-tools $COMMAND "$@"
else
  # Default to build command
  echo "Running: wow-build-tools build ${@}"
  wow-build-tools build "$@"
fi
# Capture the exit code
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "wow-build-tools failed with exit code $exit_code"
    exit $exit_code
fi

echo "wow-build-tools completed successfully"
