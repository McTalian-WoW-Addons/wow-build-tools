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
if [[ $1 == /* ]]; then
	# Remove leading slash and convert to command
	# /toc/check -> toc check
	COMMAND="${1#/}"
	COMMAND="${COMMAND//\// }"
	shift
	ARGS=("$@")
	echo "Running: wow-build-tools ${COMMAND} ${ARGS[*]}"
	wow-build-tools "${COMMAND}" "${ARGS[@]}"
else
	# Default to build command
	ARGS=("$@")
	echo "Running: wow-build-tools build ${ARGS[*]}"
	wow-build-tools build "${ARGS[@]}"
fi
# Capture the exit code
exit_code=$?

if [[ ${exit_code} -ne 0 ]]; then
	echo "wow-build-tools failed with exit code ${exit_code}"
	exit "${exit_code}"
fi

echo "wow-build-tools completed successfully"
