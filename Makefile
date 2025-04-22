# App metadata
APP_NAME := hafh-server
ENTRY := cmd/server.go
BIN_DIR := bin
CERT_DIR := certs

# Go commands
GO := go
GOFMT := go fmt

.PHONY: all build run dev fmt clean certs certs-clean

# Default target
all: build

## Build the binary
build:
	@echo "🔨 Building $(APP_NAME)..."
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(ENTRY)

## Run the app normally (no DEBUG)
run:
	@echo "🚀 Running $(APP_NAME)..."
	$(GO) run $(ENTRY) || true

## Run the app with DEBUG=true
dev:
	@echo "🧪 Running in DEBUG mode..."
	DEBUG=true $(GO) run $(ENTRY) || true

## Format Go code
fmt:
	@echo "🧼 Formatting code..."
	$(GOFMT) ./...

## Clean the bin directory and Go cache
clean:
	@echo "🧹 Cleaning up..."
	$(GO) clean
	rm -rf $(BIN_DIR)/

## Clean certs
certs-clean:
	@echo "🧨 Cleaning certs..."
	rm -rf $(CERT_DIR)/

## Generate TLS certs for mTLS
certs:
	@echo "🔐 Generating self-signed certificates for mTLS..."
	@mkdir -p $(CERT_DIR)
	# Generate CA key and cert
	openssl genrsa -out $(CERT_DIR)/ca.key 2048
	openssl req -x509 -new -nodes -key $(CERT_DIR)/ca.key -sha256 -days 365 -out $(CERT_DIR)/ca.crt -subj "/CN=HAFH-CA"

	# Generate server key and CSR
	openssl genrsa -out $(CERT_DIR)/server.key 2048
	openssl req -new -key $(CERT_DIR)/server.key -out $(CERT_DIR)/server.csr -subj "/CN=localhost"
	openssl x509 -req -in $(CERT_DIR)/server.csr -CA $(CERT_DIR)/ca.crt -CAkey $(CERT_DIR)/ca.key -CAcreateserial -out $(CERT_DIR)/server.crt -days 365 -sha256

	# Generate client key and cert
	openssl genrsa -out $(CERT_DIR)/client.key 2048
	openssl req -new -key $(CERT_DIR)/client.key -out $(CERT_DIR)/client.csr -subj "/CN=client"
	openssl x509 -req -in $(CERT_DIR)/client.csr -CA $(CERT_DIR)/ca.crt -CAkey $(CERT_DIR)/ca.key -CAcreateserial -out $(CERT_DIR)/client.crt -days 365 -sha256

	@echo "✅ Certs generated in $(CERT_DIR)/"
