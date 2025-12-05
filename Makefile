.PHONY: build run test clean deps tidy dev web-build web-dev web-install

# Variables
BINARY_NAME=m3m
BUILD_DIR=./build
MAIN_PATH=./cmd/m3m

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build frontend
web-install:
	@echo "Installing web dependencies..."
	cd web/ui && npm install

web-build:
	@echo "Building web..."
	cd web/ui && npm run build

web-dev:
	@echo "Starting web dev server..."
	cd web/ui && npm run dev

# Build the application
build: web-build
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build without web (for development)
build-backend:
	@echo "Building backend only..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME) serve

# Run in development mode (without building)
dev:
	$(GORUN) $(MAIN_PATH) serve

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	$(GOMOD) download

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Create new admin user
new-admin:
	@if [ -z "$(EMAIL)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Usage: make new-admin EMAIL=admin@example.com PASSWORD=yourpassword"; \
		exit 1; \
	fi
	$(GORUN) $(MAIN_PATH) new-admin $(EMAIL) $(PASSWORD)

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)

# Build for macOS
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PATH)

# Build all platforms
build-all: build-linux build-windows build-darwin

# Initialize project (create directories, download deps)
init:
	@echo "Initializing project..."
	@mkdir -p storage logs plugins
	$(GOMOD) tidy

# Build single plugin
build-plugin:
	@if [ -z "$(PLUGIN)" ]; then \
		echo "Usage: make build-plugin PLUGIN=telegram"; \
		exit 1; \
	fi
	@echo "Building plugin $(PLUGIN)..."
	cd plugins/$(PLUGIN) && $(GOBUILD) -buildmode=plugin -o ../$(PLUGIN).so

# Build all plugins
build-plugins:
	@echo "Building all plugins..."
	cd plugins && $(MAKE) all

# Help
help:
	@echo "M3M Makefile commands:"
	@echo "  make build         - Build frontend and backend"
	@echo "  make build-backend - Build backend only"
	@echo "  make run           - Build and run the application"
	@echo "  make dev           - Run backend in development mode"
	@echo "  make web-install   - Install web dependencies"
	@echo "  make web-build     - Build frontend"
	@echo "  make web-dev       - Run frontend dev server"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make tidy          - Tidy dependencies"
	@echo "  make init          - Initialize project"
	@echo "  make new-admin     - Create admin (EMAIL=... PASSWORD=...)"
	@echo "  make build-plugin  - Build single plugin (PLUGIN=name)"
	@echo "  make build-plugins - Build all plugins"
	@echo "  make build-all     - Build for all platforms"
