# Define the output directory and binary name
OUTPUT_DIR := dist
BINARY_NAME := wow-build-tools

# Define the build command
build:
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME)
	GOARCH=amd64 GOOS=windows go build -o $(OUTPUT_DIR)/$(BINARY_NAME).exe

# Define the clean command
clean:
	rm -rf $(OUTPUT_DIR)

# Define the run command
run: build
	./$(OUTPUT_DIR)/$(BINARY_NAME)

release:
	@cp -f ./$(OUTPUT_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@cp -f ./$(OUTPUT_DIR)/$(BINARY_NAME).exe /mnt/c/Users/robpa/bin/$(BINARY_NAME).exe

CC_THRESHOLD = 10

test:
	@mkdir -p ./.coverage
	@go test -tags="e2e" -v ./... -coverpkg=./cmd/...,./internal/... -coverprofile="./.coverage/cover.out" -json > .coverage/test-report.json || true
	@gopogh -in .coverage/test-report.json -out_html .coverage/test-report.html -out_summary .coverage/test-summary.json 2>&1 > /dev/null
	@go tool cover -html=./.coverage/cover.out -o .coverage/cover.html
	@go tool cover -func=./.coverage/cover.out > .coverage/coverage-by-function.txt
	@gocyclo -over 1 . > .coverage/complexity.txt 2>/dev/null || true
	@covreport -i .coverage/cover.out -o .coverage/report.html
	@go run ./scripts/coverage-metrics.go -threshold=$(CC_THRESHOLD) -format=json -output=.coverage/coverage-metrics.json
	@go run ./scripts/coverage-metrics.go -threshold=$(CC_THRESHOLD) -format=markdown -output=.coverage/coverage-metrics.md
	@Coverage=$$(go tool cover -func=./.coverage/cover.out | grep total | awk '{print $$3}'); \
	NumberPassed=$$(cat ./.coverage/test-summary.json | jq '.NumberOfPass'); \
	NumFailed=$$(cat ./.coverage/test-summary.json | jq '.NumberOfFail'); \
	TotalTests=$$(cat ./.coverage/test-summary.json | jq '.NumberOfTests'); \
	TotalTime=$$(cat ./.coverage/test-summary.json | jq '.TotalDuration'); \
	if [ "$$TotalTests" -eq 0 ]; then \
		echo "No tests were run. Check for compilation errors"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "Tests passed: $$NumberPassed / $$TotalTests in $$TotalTime seconds."; \
	echo "Coverage: $$Coverage"; \
	echo "Test report: file:///.coverage/test-report.html"; \
	echo "Coverage report: file:///.coverage/report.html"; \
	echo "Coverage metrics: file:///.coverage/coverage-metrics.md"; \
	echo ""; \
	go run ./scripts/coverage-metrics.go -threshold=$(CC_THRESHOLD)

# Define the default target
.PHONY: all
all: build
