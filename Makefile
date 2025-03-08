# Makefile for MIDIRouter
# =====================

# Variables
BIN_NAME := midirouter
BIN_DIR := bin
CMD_DIR := cmd/midirouter
BUILD_FLAGS :=
GOFLAGS :=

# Default target
.PHONY: all
all: build ## Build the application (default)

# Help command
.PHONY: help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build the application
.PHONY: build
build: ## Build the application
	go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BIN_NAME) $(CMD_DIR)/$(BIN_NAME).go

# Clean build artifacts
.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BIN_DIR)/$(BIN_NAME)
	go clean

# Run the application
.PHONY: run
run: build ## Build and run the application
	./$(BIN_DIR)/$(BIN_NAME)

# Format code
.PHONY: fmt
fmt: ## Format code using gofmt
	go fmt ./...

# Run go vet
.PHONY: vet
vet: ## Run go vet on the code
	go vet ./...

# Tidy dependencies
.PHONY: tidy
tidy: ## Run go mod tidy to clean up dependencies
	go mod tidy

# Check code quality (runs formatting, linting, and vet)
.PHONY: check
check: fmt vet tidy ## Run all code quality checks

# Install dependencies
.PHONY: deps
deps: ## Ensure dependencies are installed
	go mod download

# Build for specific platforms
.PHONY: build-linux
build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64 $(CMD_DIR)/$(BIN_NAME).go

.PHONY: build-macos
build-macos: ## Build for macOS
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BIN_NAME)-macos-amd64 $(CMD_DIR)/$(BIN_NAME).go

.PHONY: build-windows
build-windows: ## Build for Windows
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe $(CMD_DIR)/$(BIN_NAME).go

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-macos build-windows ## Build for all platforms

# Run tests
.PHONY: test
test: ## Run tests
	go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -cover ./...

# Generate test coverage report
.PHONY: coverage
coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Install the binary into ~/go/bin
.PHONY: install
install: build ## Install the binary into ~/go/bin
	cp ./$(BIN_DIR)/$(BIN_NAME) ~/go/bin/$(BIN_NAME)
