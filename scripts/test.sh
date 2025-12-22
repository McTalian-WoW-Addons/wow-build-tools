#!/bin/bash

CC_THRESHOLD=${CC_THRESHOLD:-10}

echo "Creating coverage directory..."
mkdir -p ./.coverage

echo "Running tests with coverage..."
set +e # Allow test command to fail so we can capture output and generate reports
go test -tags="e2e" -v ./... -coverpkg=./cmd/...,./internal/... -coverprofile="./.coverage/cover.out" >.coverage/test-output.txt 2>&1
TEST_EXIT_CODE=$?
set -e # Re-enable exit on error for the rest of the script

echo "Converting test output to JSON..."
cat .coverage/test-output.txt | go tool test2json >.coverage/test-report.json

echo "Generating test reports..."
gopogh -in .coverage/test-report.json -out_html .coverage/test-report.html -out_summary .coverage/test-summary.json 2>&1 >/dev/null

# Extract test metrics
NumberPassed=$(cat ./.coverage/test-summary.json | jq '.NumberOfPass')
NumFailed=$(cat ./.coverage/test-summary.json | jq '.NumberOfFail')
TotalTests=$(cat ./.coverage/test-summary.json | jq '.NumberOfTests')
TotalTime=$(cat ./.coverage/test-summary.json | jq '.TotalDuration')

# Check if any tests ran
if [ "$TotalTests" -eq 0 ]; then
	echo ""
	echo "❌ No tests were run. Check for compilation errors:"
	echo ""
	cat .coverage/test-output.txt
	echo ""
	echo "Full output saved to: $(pwd)/.coverage/test-output.txt"
	exit 1
fi

if [ "$TEST_EXIT_CODE" -ne 0 ]; then
	rm -f ./.coverage/cover.out
fi

# Generate coverage reports only if we have coverage data
if [ -f ./.coverage/cover.out ]; then
	echo "Generating coverage reports..."
	go tool cover -html=./.coverage/cover.out -o .coverage/cover.html
	go tool cover -func=./.coverage/cover.out >.coverage/coverage-by-function.txt
	gocyclo -over 1 . >.coverage/complexity.txt 2>/dev/null || true
	covreport -i .coverage/cover.out -o .coverage/report.html
	go run ./scripts/coverage-metrics.go -threshold=$CC_THRESHOLD -format=json -output=.coverage/coverage-metrics.json
	go run ./scripts/coverage-metrics.go -threshold=$CC_THRESHOLD -format=markdown -output=.coverage/coverage-metrics.md

	Coverage=$(go tool cover -func=./.coverage/cover.out | grep total | awk '{print $3}')
else
	echo "⚠️  No coverage data generated"
	Coverage="N/A"
fi

# Display results
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Tests passed:    $NumberPassed / $TotalTests"
echo "Tests failed:    $NumFailed"
echo "Duration:        $TotalTime seconds"
if [ -f ./.coverage/cover.out ]; then
	echo "Coverage:        $Coverage"
fi
echo ""
echo "Reports:"
echo "  Test report:     file://$(pwd)/.coverage/test-report.html"
if [ -f ./.coverage/cover.out ]; then
	echo "  Coverage report: file://$(pwd)/.coverage/report.html"
	echo "  Coverage metrics: file://$(pwd)/.coverage/coverage-metrics.md"
fi
echo "  Test output:     file://$(pwd)/.coverage/test-output.txt"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Run coverage metrics check if we have coverage
if [ -f ./.coverage/cover.out ]; then
	go run ./scripts/coverage-metrics.go -threshold=$CC_THRESHOLD
fi

# Check for failures and show relevant output
if [ "$NumFailed" -ne 0 ]; then
	echo "❌ $NumFailed test(s) failed"
	echo ""
	echo "Failed test output:"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	# Show failure lines
	grep -B 5 -A 15 "^FAIL" .coverage/test-output.txt
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	echo ""
	echo "Full output: $(pwd)/.coverage/test-output.txt"
	exit 1
fi

if [ $TEST_EXIT_CODE -ne 0 ]; then
	echo "❌ Test command failed with exit code $TEST_EXIT_CODE"
	echo ""
	# Check if it's a build failure (compilation error)
	if grep -q "\[build failed\]" .coverage/test-output.txt; then
		echo "Compilation errors:"
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
		# Show package header and compilation errors (only non-indented lines)
		grep -E "^# .+\.test\]$|^[^ ].*\.go:[0-9]+:" .coverage/test-output.txt
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	else
		echo "Test output:"
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
		tail -100 .coverage/test-output.txt
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	fi
	echo ""
	echo "Full output: $(pwd)/.coverage/test-output.txt"
	exit $TEST_EXIT_CODE
fi

echo "✅ All tests passed!"
