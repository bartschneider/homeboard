# E-Paper Dashboard Makefile

.PHONY: build run test clean install dev help

# Build configuration
APP_NAME := homeboard
BUILD_DIR := bin
GO_FILES := $(shell find . -name "*.go" -type f)
BINARY := $(BUILD_DIR)/$(APP_NAME)

# Default target
help:
	@echo "E-Paper Dashboard Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Build the Go server binary"
	@echo "  run      - Build and run the server"
	@echo "  dev      - Run in development mode with hot reload"
	@echo "  test     - Run tests and validate widgets"
	@echo "  install  - Install Python dependencies"
	@echo "  clean    - Clean build artifacts"
	@echo "  package  - Create deployment package"
	@echo "  help     - Show this help message"

# Build the application
build: $(BINARY)

$(BINARY): $(GO_FILES)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BINARY) cmd/server/main.go
	@echo "Build complete: $(BINARY)"

# Run the application
run: build
	@echo "Starting E-Paper Dashboard..."
	./$(BINARY) -config config.json -verbose

# Development mode with file watching (requires entr)
dev:
	@echo "Starting development mode..."
	@echo "Note: Install 'entr' for auto-reload: brew install entr"
	find . -name "*.go" | entr -r make run

# Test the application
test: build
	@echo "Running tests..."
	go test ./...
	@echo "Testing widgets..."
	python3 widgets/clock.py '{"format": "%H:%M", "timezone": "Local"}'
	python3 widgets/system.py '{"show_cpu": true}'
	python3 widgets/todo.py '{"use_static": true, "max_items": 3}'
	@echo "Testing server endpoints..."
	./$(BINARY) -config config.json &
	sleep 2
	curl -f http://localhost:8081/health || (echo "Health check failed" && exit 1)
	curl -f http://localhost:8081/api/config || (echo "Config API failed" && exit 1)
	pkill -f $(BINARY) || true
	@echo "All tests passed!"

# Install Python dependencies
install:
	@echo "Installing Python dependencies..."
	pip3 install psutil pytz requests
	@echo "Verifying installation..."
	python3 -c "import psutil, pytz; print('Core dependencies installed')"
	python3 -c "import requests; print('Requests installed')" 2>/dev/null || echo "Note: requests library for weather widget not installed"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f dashboard.log
	go clean
	@echo "Clean complete"

# Create deployment package
package: build
	@echo "Creating deployment package..."
	@VERSION=$$(git describe --tags --always 2>/dev/null || echo "dev")
	@PACKAGE_NAME="homeboard-$$VERSION"
	@mkdir -p "$$PACKAGE_NAME"
	cp $(BINARY) "$$PACKAGE_NAME/"
	cp config.json "$$PACKAGE_NAME/"
	cp -r widgets/ "$$PACKAGE_NAME/"
	cp -r kindle-extension/ "$$PACKAGE_NAME/"
	cp README.md INSTALL.md "$$PACKAGE_NAME/"
	tar -czf "$$PACKAGE_NAME.tar.gz" "$$PACKAGE_NAME"
	rm -rf "$$PACKAGE_NAME"
	@echo "Package created: $$PACKAGE_NAME.tar.gz"

# Quick deployment test
deploy-test: build
	@echo "Testing deployment..."
	@mkdir -p test-deploy
	cp $(BINARY) config.json test-deploy/
	cp -r widgets/ test-deploy/
	cd test-deploy && ./$(APP_NAME) -config config.json &
	sleep 2
	curl -f http://localhost:8081/health
	pkill -f test-deploy/$(APP_NAME) || true
	rm -rf test-deploy
	@echo "Deployment test successful!"

# Kindle extension package
kindle-package:
	@echo "Creating Kindle extension package..."
	@mkdir -p kindle-dashboard-launcher
	cp kindle-extension/* kindle-dashboard-launcher/
	zip -r kindle-dashboard-launcher.zip kindle-dashboard-launcher/
	rm -rf kindle-dashboard-launcher
	@echo "Kindle extension package: kindle-dashboard-launcher.zip"

# Development server with custom config
dev-config:
	@echo "Starting with extended config..."
	./$(BINARY) -config config-extended.json -verbose

# Performance test
perf-test: build
	@echo "Running performance tests..."
	@echo "Testing widget execution speed..."
	time python3 widgets/clock.py '{}'
	time python3 widgets/system.py '{}'
	time python3 widgets/todo.py '{"use_static": true}'
	@echo "Testing server response time..."
	./$(BINARY) -config config.json &
	sleep 2
	@echo "Warming up..."
	curl -s http://localhost:8081/ > /dev/null
	@echo "Testing response time..."
	time curl -s http://localhost:8081/ > /dev/null
	pkill -f $(BINARY) || true
	@echo "Performance test complete"

# Widget validation
validate-widgets:
	@echo "Validating all widgets..."
	@for widget in widgets/*.py; do \
		echo "Testing $$widget..."; \
		python3 "$$widget" '{}' > /dev/null || echo "$$widget failed"; \
	done
	@echo "Widget validation complete"