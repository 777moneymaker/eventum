# Go settings
GO = go
GOFMT = gofmt
GOTEST = go test
GOLINT = go vet

# Directories
SRC_DIR = ./cmd
BIN_DIR = ./bin
DEP_DIR = ./vendor

# Go files
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Target - the default target is eventum, but you can specify with `make build target=eventum`
TARGET ?= eventum

# Dependencies
deps:
	@echo "Installing dependencies..."
	@$(GO) mod tidy
	@$(GO) mod vendor
	@echo "Dependencies installed."

# Format the code
fmt:
	@echo "Formatting Go code..."
	@$(GOFMT) -w $(GO_FILES)
	@echo "Code formatting complete."

# Build the project - accepts target
build:
	@echo "Building $(TARGET)..."
	@$(GO) build -o $(BIN_DIR)/$(TARGET) $(SRC_DIR)/$(TARGET)/main.go
	@echo "Build complete: $(BIN_DIR)/$(TARGET)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR) $(DEP_DIR)
	@echo "Clean complete."


# Run the target (build and execute)
run: build
	@echo "Running $(TARGET)..."
	@$(BIN_DIR)/$(TARGET)
	@echo "$(TARGET) executed."

# Default target
.PHONY: build fmt lint test clean deps install
