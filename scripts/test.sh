#!/bin/bash

CC_THRESHOLD=${CC_THRESHOLD:-10}

function create_coverage_dir {
	echo "Creating coverage directory..."
	mkdir -p ./.coverage

	abs_coverage_dir=$(realpath ./.coverage)
}

function run_tests {
	echo "Running tests with coverage..."
	set +e # Allow test command to fail so we can capture output and generate reports
	go test -tags="e2e" -v ./... -coverpkg=./cmd/...,./internal/... -coverprofile="./.coverage/cover.out" >.coverage/test-output.txt 2>&1
	TEST_EXIT_CODE=$?
	set -e # Re-enable exit on error for the rest of the script
}

function convert_results_to_json {
	echo "Convert test output to JSON..."
	go tool test2json <.coverage/test-output.txt >.coverage/test-report.json
}

function generate_test_report {
	echo "Generating pass/fail report..."
	gopogh -in .coverage/test-report.json -out_html .coverage/test-report.html -out_summary .coverage/test-summary.json >/dev/null 2>&1
}

function extract_test_metrics {
	# Extract test metrics
	NumberPassed=$(jq '.NumberOfPass' <./.coverage/test-summary.json)
	NumFailed=$(jq '.NumberOfFail' <./.coverage/test-summary.json)
	TotalTests=$(jq '.NumberOfTests' <./.coverage/test-summary.json)
	TotalTime=$(jq '.TotalDuration' <./.coverage/test-summary.json)
}

function validate_test_execution {
	# Check if any tests ran
	if [[ ${TotalTests} -eq 0 ]]; then
		echo ""
		echo "❌ No tests were run. Check for compilation errors:"
		echo ""
		cat .coverage/test-output.txt
		echo ""
		echo "Full output saved to: ${abs_coverage_dir}/test-output.txt"
		exit 1
	fi

	if [[ ${TEST_EXIT_CODE} -ne 0 ]]; then
		# If tests failed, remove coverage data to avoid misleading reports
		rm -f ./.coverage/cover.out
	fi
}

function generate_coverage_report {
	# Generate coverage reports only if we have coverage data
	if [[ -f ./.coverage/cover.out ]]; then
		echo "Generating coverage reports..."
		go tool cover -html=./.coverage/cover.out -o .coverage/cover.html
		go tool cover -func=./.coverage/cover.out >.coverage/coverage-by-function.txt
		gocyclo -over 1 . >.coverage/complexity.txt 2>/dev/null || true
		covreport -i .coverage/cover.out -o .coverage/report.html
		go run ./scripts/coverage-metrics.go -threshold="${CC_THRESHOLD}" -format=json -output=.coverage/coverage-metrics.json
		go run ./scripts/coverage-metrics.go -threshold="${CC_THRESHOLD}" -format=markdown -output=.coverage/coverage-metrics.md

		total_line=$(grep total ./.coverage/coverage-by-function.txt)
		Coverage=$(awk '{print $3}' <<<"${total_line}")
	else
		echo "⚠️  No coverage data generated"
		Coverage="N/A"
	fi
}

function display_test_results {
	echo ""
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	echo "Test Results"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	echo "Tests passed:    ${NumberPassed} / ${TotalTests}"
	echo "Tests failed:    ${NumFailed}"
	echo "Duration:        ${TotalTime} seconds"
	if [[ -f ./.coverage/cover.out ]]; then
		echo "Coverage:        ${Coverage}"
	fi
	echo ""
	echo "Reports:"
	echo "  Test report:     file://${abs_coverage_dir}/test-report.html"
	if [[ -f ./.coverage/cover.out ]]; then
		echo "  Coverage report: file://${abs_coverage_dir}/report.html"
		echo "  Coverage metrics: file://${abs_coverage_dir}/coverage-metrics.md"
	fi
	echo "  Test output:     file://${abs_coverage_dir}/test-output.txt"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	echo ""

	# Check for failures and show relevant output
	if [[ ${NumFailed} -ne 0 ]]; then
		echo "${NumFailed} test(s) failed"
		echo ""
		echo "Failed test output:"
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
		# Show failure lines
		grep -B 5 -A 15 "^FAIL" .coverage/test-output.txt
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
		echo ""
		echo "Full output: ${abs_coverage_dir}/test-output.txt"
		exit 1
	fi

	if [[ ${TEST_EXIT_CODE} -ne 0 ]]; then
		echo "❌ Test command failed with exit code ${TEST_EXIT_CODE}"
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
		echo "Full output: ${abs_coverage_dir}/test-output.txt"
		exit "${TEST_EXIT_CODE}"
	fi
}

function display_results {
	display_test_results

	# Run coverage metrics check if we have coverage
	if [[ -f ./.coverage/cover.out ]]; then
		go run ./scripts/coverage-metrics.go -threshold="${CC_THRESHOLD}"
	fi

	echo "✅ All tests passed!"
}

function main {
	create_coverage_dir
	run_tests
	convert_results_to_json
	generate_test_report
	extract_test_metrics
	validate_test_execution
	generate_coverage_report
	display_results
}

main
