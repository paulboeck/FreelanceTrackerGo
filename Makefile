.PHONY: help build build-release test test-verbose test-coverage test-ci test-e2e test-unit test-http clean run migrate generate install-tools tidy docker-build docker-run \
	build-macos build-macos-intel build-macos-arm build-macos-universal build-macos-app \
	build-windows build-windows-amd64 build-windows-arm64 \
	build-linux build-linux-amd64 build-linux-arm64 build-linux-arm \
	build-all-platforms build-all-specific package-macos-app package-windows package-linux package-all

# Default target
.DEFAULT_GOAL := help

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Build directories
BUILD_DIR := bin
DIST_DIR := dist
BINARY_NAME := web
APP_NAME := FreelanceTracker

# Platform-specific settings
MACOS_BINARY := $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
MACOS_ARM_BINARY := $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
MACOS_UNIVERSAL_BINARY := $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal
WINDOWS_BINARY := $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
LINUX_BINARY := $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/web

build-release: ## Build optimized release binary with version info
	@echo "Building release $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/web
	@echo "Built $(BINARY_NAME) $(VERSION) at $(BUILD_TIME)"

test: ## Run all tests
	@echo "Running tests..."
	go test ./...

test-verbose: ## Run all tests with verbose output
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-ci: ## Run tests with CI settings (race detector, coverage)
	@echo "Running tests with CI configuration..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Tests completed. Coverage report: coverage.out"

test-e2e: ## Run only end-to-end browser tests
	@echo "Running e2e tests..."
	go test -v -run "TestE2E" ./cmd/web -timeout=60s

test-unit: ## Run only unit tests (exclude e2e and HTTP integration tests)
	@echo "Running unit tests..."
	go test -v -short ./internal/...

test-http: ## Run only HTTP integration tests
	@echo "Running HTTP integration tests..."
	go test -v -run "Test.*Handler" ./cmd/web -timeout=30s

clean: ## Clean build artifacts and test databases
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "Cleaning test databases..."
	find . -name "test_*.db" -type f -delete
	@echo "Killing any orphaned test processes..."
	-pkill -f "web -addr" 2>/dev/null || true

# Platform-specific builds

# macOS builds
build-macos-intel: ## Build for macOS Intel (x86_64) only
	@echo "Building macOS Intel binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -trimpath $(LDFLAGS) -o $(MACOS_BINARY) ./cmd/web
	@echo "Built macOS Intel binary: $(MACOS_BINARY)"

build-macos-arm: ## Build for macOS Apple Silicon (ARM64) only
	@echo "Building macOS ARM64 binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -trimpath $(LDFLAGS) -o $(MACOS_ARM_BINARY) ./cmd/web
	@echo "Built macOS ARM64 binary: $(MACOS_ARM_BINARY)"

build-macos-universal: build-macos-intel build-macos-arm ## Build macOS universal binary (Intel + Apple Silicon)
	@echo "Creating universal binary..."
	lipo -create -output $(MACOS_UNIVERSAL_BINARY) $(MACOS_BINARY) $(MACOS_ARM_BINARY)
	@echo "Built universal macOS binary: $(MACOS_UNIVERSAL_BINARY)"

build-macos: build-macos-universal ## Build for macOS (creates universal binary by default)

build-macos-app: build-macos ## Build macOS .app bundle
	@echo "Creating macOS application bundle..."
	@./scripts/build-macos-app.sh $(VERSION)
	@echo "macOS application created: $(DIST_DIR)/$(APP_NAME).app"

# Windows builds
build-windows-amd64: ## Build for Windows x86_64 (64-bit Intel/AMD)
	@echo "Building Windows x86_64 binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -trimpath $(LDFLAGS) -o $(WINDOWS_BINARY) ./cmd/web
	@echo "Built Windows x86_64 binary: $(WINDOWS_BINARY)"

build-windows-arm64: ## Build for Windows ARM64
	@echo "Building Windows ARM64 binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-w64-mingw32-gcc go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/web
	@echo "Built Windows ARM64 binary: $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe"

build-windows: build-windows-amd64 ## Build for Windows (x86_64 by default)

# Linux builds
build-linux-amd64: ## Build for Linux x86_64 (64-bit Intel/AMD)
	@echo "Building Linux x86_64 binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -trimpath $(LDFLAGS) -o $(LINUX_BINARY) ./cmd/web
	@echo "Built Linux x86_64 binary: $(LINUX_BINARY)"

build-linux-arm64: ## Build for Linux ARM64 (e.g., Raspberry Pi, AWS Graviton)
	@echo "Building Linux ARM64 binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/web
	@echo "Built Linux ARM64 binary: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

build-linux-arm: ## Build for Linux ARM (32-bit, e.g., older Raspberry Pi)
	@echo "Building Linux ARM binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm ./cmd/web
	@echo "Built Linux ARM binary: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm"

build-linux: build-linux-amd64 ## Build for Linux (x86_64 by default)

build-all-platforms: build-macos build-windows build-linux docker-build ## Build for all platforms (macOS, Windows, Linux, Docker)
	@echo "All platform builds completed!"
	@echo "Binaries created:"
	@ls -lh $(BUILD_DIR)/
	@echo ""
	@echo "Docker image: freelance-tracker:$(VERSION)"

build-all-specific: build-macos-intel build-macos-arm build-windows-amd64 build-linux-amd64 build-linux-arm64 ## Build all platform-specific optimized binaries
	@echo ""
	@echo "All platform-specific builds completed!"
	@echo ""
	@echo "macOS Intel (x86_64):    $(MACOS_BINARY)"
	@echo "macOS ARM64 (M1/M2/M3):  $(MACOS_ARM_BINARY)"
	@echo "Windows x86_64:          $(WINDOWS_BINARY)"
	@echo "Linux x86_64:            $(LINUX_BINARY)"
	@echo "Linux ARM64:             $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"
	@echo ""
	@echo "File sizes:"
	@ls -lh $(BUILD_DIR)/ | grep web-

package-macos-app: build-macos-app ## Package macOS .app into a DMG
	@echo "Creating macOS DMG..."
	@hdiutil create -volname "$(APP_NAME)" -srcfolder $(DIST_DIR)/$(APP_NAME).app -ov -format UDZO $(DIST_DIR)/$(APP_NAME)-$(VERSION).dmg
	@echo "✓ DMG created: $(DIST_DIR)/$(APP_NAME)-$(VERSION).dmg"

package-windows: build-windows ## Package Windows executable with resources
	@echo "Creating Windows package..."
	@./scripts/build-windows.sh $(VERSION)
	@echo "Creating ZIP archive..."
	@cd $(DIST_DIR) && zip -r $(APP_NAME)-$(VERSION)-windows.zip windows/
	@echo "✓ Windows package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows.zip"

package-linux: build-linux ## Package Linux binary with resources
	@echo "Creating Linux package..."
	@mkdir -p $(DIST_DIR)/linux
	@cp $(LINUX_BINARY) $(DIST_DIR)/linux/freelance-tracker
	@cp -R ui $(DIST_DIR)/linux/
	@cp -R migrations $(DIST_DIR)/linux/
	@echo "Creating tarball..."
	@cd $(DIST_DIR) && tar czf $(APP_NAME)-$(VERSION)-linux-amd64.tar.gz linux/
	@echo "✓ Linux package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

package-all: package-macos-app package-windows package-linux ## Create distribution packages for all platforms
	@echo ""
	@echo "All packages created:"
	@ls -lh $(DIST_DIR)/*.{dmg,zip,tar.gz} 2>/dev/null || true

run: ## Run the application (default: localhost:8080)
	@echo "Running application..."
	go run ./cmd/web

run-dev: ## Run the application with development settings
	@echo "Running application in development mode..."
	go run ./cmd/web -addr=":8080" -dsn="./freelance_tracker_dev.db"

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 ./freelance_tracker.db up

migrate-status: ## Show migration status
	@echo "Migration status:"
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 ./freelance_tracker.db status

generate: ## Generate code from SQL queries using sqlc
	@echo "Generating code with sqlc..."
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

install-tools: ## Install development tools (goose, sqlc)
	@echo "Installing development tools..."
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Tools installed successfully"

tidy: ## Tidy and verify go.mod dependencies
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t freelance-tracker:$(VERSION) -t freelance-tracker:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/data freelance-tracker:latest

docker-compose-up: ## Start application using docker-compose
	docker-compose up -d

docker-compose-down: ## Stop application using docker-compose
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

all: clean tidy generate test build ## Run full build pipeline (clean, tidy, generate, test, build)
