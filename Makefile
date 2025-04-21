# App metadata
APP_NAME := hafh-server
ENTRY := cmd/server.go
BIN_DIR := bin

# Go commands
GO := go
GOFMT := go fmt

.PHONY: all build run dev fmt clean

# Default target
all: build

## Build the binary
build:
	@echo "🔨 Building $(APP_NAME)..."
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(ENTRY)

## Run the app normally (no DEBUG)
run:
	@echo "🚀 Running $(APP_NAME)..."
	$(GO) run $(ENTRY)

## Run the app with DEBUG=true
dev:
	@echo "🧪 Running in DEBUG mode..."
	DEBUG=true $(GO) run $(ENTRY)

## Format Go code
fmt:
	@echo "🧼 Formatting code..."
	$(GOFMT) ./...

## Clean the bin directory and Go cache
clean:
	@echo "🧹 Cleaning up..."
	$(GO) clean
	rm -rf $(BIN_DIR)/
