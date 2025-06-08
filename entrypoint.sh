#!/bin/bash
set -e

# GitHub Actions sets the workspace to /github/workspace
cd /github/workspace

# Ensure we're running in the GitHub Actions environment
export CI="true"
export GITHUB_ACTIONS="true"

# Get arguments passed from action.yml args section
# $1 contains the args input from the action
ARGS="$1"

# Run wow-build-tools with the provided arguments
echo "Running: wow-build-tools build ${ARGS}"
wow-build-tools build ${ARGS}

# Capture the exit code
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "wow-build-tools failed with exit code $exit_code"
    exit $exit_code
fi

echo "wow-build-tools completed successfully"
