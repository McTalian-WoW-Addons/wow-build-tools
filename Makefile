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

release: build
	@cp -f ./$(OUTPUT_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@cp -f ./$(OUTPUT_DIR)/$(BINARY_NAME).exe /mnt/c/Users/robpa/bin/$(BINARY_NAME).exe

tools:
	@echo "Installing Go tools..."
	@cat go.tools | xargs -n1 go install -v
	@echo "Checking for jq (required by scripts/test.sh)..."
	@if command -v jq >/dev/null 2>&1; then \
		echo "✓ jq is already installed"; \
		echo "✓ Tools installed"; \
	else \
		echo "✗ jq not found on PATH."; \
		echo "  jq is required by scripts/test.sh."; \
		echo "  Please install jq using your package manager, for example:"; \
		echo "    macOS (Homebrew):   brew install jq"; \
		echo "    Debian/Ubuntu:      sudo apt-get update && sudo apt-get install -y jq"; \
		echo "    Fedora/RHEL/CentOS: sudo dnf install -y jq"; \
		echo "    Arch Linux:         sudo pacman -S jq"; \
		echo "    Windows (choco):    choco install jq"; \
		echo "    Windows (scoop):    scoop install jq"; \
		exit 1; \
	fi

CC_THRESHOLD = 10

test:
	@CC_THRESHOLD=$(CC_THRESHOLD) ./scripts/test.sh

# Define the default target
.PHONY: all
all: build
