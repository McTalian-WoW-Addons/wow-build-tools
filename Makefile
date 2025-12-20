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

test:
	@mkdir -p ./.coverage
	@go test -tags="e2e" -v ./... -coverpkg=./... -coverprofile="./.coverage/cover.out" -json > .coverage/test-report.json || true
	@gopogh -in .coverage/test-report.json -out_html .coverage/test-report.html -out_summary .coverage/test-summary.json 2>&1 > /dev/null
	@go tool cover -html=./.coverage/cover.out -o .coverage/cover.html
	@covreport -i .coverage/cover.out -o .coverage/report.html
	@echo "Test report: file:///.coverage/test-report.html";
	@echo "Coverage report: file:///.coverage/report.html";

# Define the default target
.PHONY: all
all: build