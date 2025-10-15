# Makefile for UniFi Threat Sync

# Variables
BINARY_NAME=unifi-threat-sync
MAIN_PATH=./cmd/$(BINARY_NAME)
BUILD_DIR=bin
DOCKER_IMAGE=ghcr.io/0x4272616e646f6e/$(BINARY_NAME)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Detect OS for different commands
UNAME_S := $(shell uname -s)

.PHONY: all build clean test coverage lint fmt vet deps tidy run help
.PHONY: docker-build docker-push docker-run docker-clean
.PHONY: install uninstall release

## Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "UniFi Threat Sync - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':'

## all: Build the application
all: clean deps build

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux (amd64)
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

## build-darwin: Build for macOS (amd64 and arm64)
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

## build-windows: Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All builds complete!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## test: Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -timeout 30s ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: brew install golangci-lint" && exit 1)
	@golangci-lint run --timeout 5m ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy
	@$(GOMOD) verify

## run: Run the application locally
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

## install: Install the binary to GOBIN
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)

## uninstall: Remove installed binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built: $(DOCKER_IMAGE):$(VERSION)"

## docker-build-no-cache: Build Docker image without cache
docker-build-no-cache:
	@echo "Building Docker image (no cache)..."
	@docker build --no-cache -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

## docker-push: Push Docker image to registry
docker-push: docker-build
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

## docker-run: Run Docker container locally
docker-run:
	@echo "Running Docker container..."
	@docker run --rm -it \
		--name $(BINARY_NAME) \
		-e UNIFI_URL=$${UNIFI_URL} \
		-e UNIFI_USER=$${UNIFI_USER} \
		-e UNIFI_PASS=$${UNIFI_PASS} \
		-e FEED_URLS=$${FEED_URLS} \
		--read-only \
		--security-opt=no-new-privileges:true \
		$(DOCKER_IMAGE):latest

## docker-run-config: Run Docker container with config file
docker-run-config:
	@echo "Running Docker container with config..."
	@docker run --rm -it \
		--name $(BINARY_NAME) \
		-v $(PWD)/configs/config.yaml:/config/config.yaml:ro \
		--read-only \
		--security-opt=no-new-privileges:true \
		$(DOCKER_IMAGE):latest

## docker-clean: Remove Docker images
docker-clean:
	@echo "Removing Docker images..."
	@docker rmi $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest 2>/dev/null || true

## docker-scan: Scan Docker image for vulnerabilities
docker-scan:
	@echo "Scanning Docker image..."
	@docker scout cves $(DOCKER_IMAGE):latest || docker scan $(DOCKER_IMAGE):latest || echo "Install docker scout or trivy for scanning"

## release: Create a release build
release: clean deps test lint build-all
	@echo "Release $(VERSION) ready!"
	@ls -lh $(BUILD_DIR)/

## version: Display version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"
