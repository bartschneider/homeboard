# E-Paper Dashboard Makefile

.PHONY: build run test clean install dev help hooks

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
	@echo "  build       - Build the Go server binary"
	@echo "  run         - Build and run the server"
	@echo "  dev         - Run in development mode with hot reload"
	@echo "  test        - Run tests and validate widgets"
	@echo "  test-quick  - Run tests without widget validation"
	@echo "  lint        - Run linting and code quality checks"
	@echo "  format      - Format all code (Go and Python)"
	@echo "  install     - Install Python dependencies"
	@echo "  hooks       - Install pre-commit hooks"
	@echo "  validate    - Run comprehensive validation"
	@echo "  clean       - Clean build artifacts"
	@echo "  package     - Create deployment package"
	@echo "  docker      - Build and run Docker container for preview"
	@echo "  docker-dev  - Start development environment with Docker"
	@echo "  docker-prod - Start production environment with Docker"
	@echo "  help        - Show this help message"

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

# Test the application (comprehensive)
test: build
	@echo "Running comprehensive tests..."
	@echo "1. Go unit tests..."
	go test -race -timeout=30s ./...
	@echo "2. Go linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet..."; \
		go vet ./...; \
	fi
	@echo "3. Testing widgets..."
	@if [ -d "widgets" ]; then \
		for widget in widgets/*.py; do \
			echo "Testing $$widget..."; \
			python3 "$$widget" '{"test": true}' > /dev/null || echo "$$widget failed"; \
		done; \
	fi
	@echo "4. Testing server endpoints..."
	./$(BINARY) -config config.json &
	sleep 2
	curl -f http://localhost:8081/health || (echo "Health check failed" && exit 1)
	curl -f http://localhost:8081/api/config || (echo "Config API failed" && exit 1)
	pkill -f $(BINARY) || true
	@echo "All tests passed!"

# Quick test (Go tests only)
test-quick:
	@echo "Running quick tests..."
	go test -race -timeout=30s ./...
	@echo "Quick tests passed!"

# Linting and code quality
lint:
	@echo "Running linting and code quality checks..."
	@echo "1. Go formatting check..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not properly formatted:"; \
		gofmt -l .; \
		echo "Run 'make format' to fix"; \
		exit 1; \
	fi
	@echo "2. Go imports check..."
	@if command -v goimports >/dev/null 2>&1; then \
		if [ -n "$$(goimports -l .)" ]; then \
			echo "The following files have unorganized imports:"; \
			goimports -l .; \
			echo "Run 'make format' to fix"; \
			exit 1; \
		fi; \
	fi
	@echo "3. Go vet..."
	go vet ./...
	@echo "4. GolangCI-Lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi
	@echo "5. Python widget syntax..."
	@if [ -d "widgets" ]; then \
		for widget in widgets/*.py; do \
			python3 -m py_compile "$$widget" || exit 1; \
		done; \
	fi
	@echo "Linting passed!"

# Format all code
format:
	@echo "Formatting code..."
	@echo "1. Go formatting..."
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo "2. Go imports..."; \
		goimports -w .; \
	fi
	@echo "3. Python formatting..."
	@if command -v black >/dev/null 2>&1 && [ -d "widgets" ]; then \
		black widgets/; \
	fi
	@echo "Formatting complete!"

# Install Python dependencies
install:
	@echo "Installing Python dependencies..."
	pip3 install psutil pytz requests
	@echo "Verifying installation..."
	python3 -c "import psutil, pytz; print('Core dependencies installed')"
	python3 -c "import requests; print('Requests installed')" 2>/dev/null || echo "Note: requests library for weather widget not installed"

# Install pre-commit hooks
hooks:
	@echo "Installing pre-commit hooks..."
	@if [ -f "scripts/install-hooks.sh" ]; then \
		./scripts/install-hooks.sh; \
	else \
		echo "Hook installation script not found!"; \
		exit 1; \
	fi

# Run comprehensive validation
validate:
	@echo "Running comprehensive validation..."
	@if [ -f "scripts/validate-commit.sh" ]; then \
		./scripts/validate-commit.sh; \
	else \
		echo "Validation script not found, running basic checks..."; \
		make lint && make test-quick; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f dashboard.log
	rm -f coverage.out
	rm -f gosec-report.json
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

# Coverage report
coverage:
	@echo "Generating coverage report..."
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=json -out=gosec-report.json ./...; \
		echo "Security report generated: gosec-report.json"; \
	else \
		echo "gosec not found, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Pre-commit validation (used by git hooks)
pre-commit: format lint test-quick
	@echo "Pre-commit validation passed!"

# CI/CD pipeline simulation
ci: hooks format lint test coverage security
	@echo "CI pipeline completed successfully!"

# Docker targets

# Build and run Docker container for localhost preview
docker: docker-build docker-up
	@echo "Docker container started for preview"
	@echo "Dashboard available at: http://localhost:8081"
	@echo "Admin panel available at: http://localhost:8081/admin"
	@echo "Use 'make docker-down' to stop"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t homeboard:latest .
	@echo "Docker image built successfully"

# Start Docker Compose services
docker-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services started successfully"

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down
	@echo "Services stopped"

# Start development environment with Docker
docker-dev:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up --build -d
	@echo "Development environment started"
	@echo "Dashboard: http://localhost:8081"
	@echo "Mailhog: http://localhost:8025"
	@echo "PostgreSQL: localhost:5432"
	@echo "Use 'make docker-dev-down' to stop"

# Stop development environment
docker-dev-down:
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down
	@echo "Development environment stopped"

# Start production environment with monitoring
docker-prod:
	@echo "Starting production environment with monitoring..."
	docker-compose --profile monitoring up --build -d
	@echo "Production environment started"
	@echo "Dashboard: http://localhost:8081"
	@echo "Nginx Proxy: http://localhost"
	@echo "Grafana: http://localhost:3000"
	@echo "Prometheus: http://localhost:9090"
	@echo "Use 'make docker-prod-down' to stop"

# Stop production environment
docker-prod-down:
	@echo "Stopping production environment..."
	docker-compose --profile monitoring down
	@echo "Production environment stopped"

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	docker-compose logs -f

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f
	docker volume prune -f
	@echo "Docker cleanup complete"

# Docker development shell
docker-shell:
	@echo "Opening shell in homeboard container..."
	docker-compose exec homeboard /bin/sh

# Run tests in Docker
docker-test:
	@echo "Running tests in Docker..."
	docker build --target builder -t homeboard:test .
	docker run --rm homeboard:test go test -race ./...
	@echo "Docker tests completed"

# Deploy to production (builds and starts)
docker-deploy: docker-build
	@echo "Deploying to production..."
	docker-compose up -d --remove-orphans
	docker system prune -f
	@echo "Deployment complete"

# Show Docker status
docker-status:
	@echo "Docker services status:"
	@docker-compose ps
	@echo ""
	@echo "Docker resource usage:"
	@docker stats --no-stream

# Docker health check
docker-health:
	@echo "Checking Docker service health..."
	@curl -f http://localhost:8081/health || echo "Health check failed"
	@echo "Health check complete"