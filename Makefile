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

tools:
	@echo "Installing Go tools..."
	@cat go.tools | xargs -n1 go install -v
	@echo "✓ Tools installed"

CC_THRESHOLD = 10

test:
	@CC_THRESHOLD=$(CC_THRESHOLD) ./scripts/test.sh

# Define the default target
.PHONY: all
all: build
