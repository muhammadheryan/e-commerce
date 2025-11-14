# Makefile untuk URL Shortener Project

# Variables
APP_NAME=url-shortener
BIN_DIR=bin
MAIN_FILE=./cmd/main.go

# Default target (yang pertama akan dijalankan saat `make`)
.PHONY: help
help: ## Menampilkan bantuan
	@echo "Available commands:"
	@echo "  make help     - Tampilkan bantuan ini"
	@echo "  make build    - Build aplikasi"
	@echo "  make test     - Jalankan test"
	@echo "  make clean    - Hapus file build"
	@echo "  make run      - Build dan jalankan aplikasi"
	@echo ""
	@echo "Development commands:"
	@echo "  make deps     - Download dependencies"
	@echo "  make mocks    - Generate mocks"
	@echo "  make fmt      - Format code"
	@echo "  make swagger  - Generate Swagger documentation"
	@echo "  make dev      - Quick dev workflow"
	@echo "  make all      - Clean + setup + test + build"
	@echo ""
	@echo "Mock commands:"
	@echo "  make mocks-url        - Generate URLRepository mock only"
	@echo "  make mocks-all        - Generate all repository mocks"
	@echo "  make mocks-everything - Generate mocks for all layers"
	@echo ""
	@echo "Database commands:"
	@echo "  make migrate-up      - Run migrations up"
	@echo "  make migrate-down    - Run migrations down"
	@echo "  make migrate-status  - Check migration status"

# Clean build artifacts
.PHONY: clean
clean: ## Hapus file build dan temporary files
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)/
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Format code
.PHONY: fmt
fmt: ## Format semua Go code
	@echo "Formatting code..."
	go fmt ./...

# Download dependencies
.PHONY: deps
deps: ## Download dan update dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build aplikasi
.PHONY: build
build: ## Build aplikasi ke folder bin/
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete! Binary: $(BIN_DIR)/$(APP_NAME)"

# Build dan jalankan aplikasi
.PHONY: run
run: build ## Build dan jalankan aplikasi
	@echo "Starting $(APP_NAME)..."
	./$(BIN_DIR)/$(APP_NAME)


## ---------- ## Testing Commands ## ---------- #

# Jalankan test
.PHONY: test
test: ## Jalankan semua test
	@echo "Running tests..."
	go test ./... -v

# Jalankan test dengan coverage
.PHONY: test-coverage
test-coverage: ## Jalankan test dengan coverage report
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"


# Generate Swagger documentation
.PHONY: swagger
swagger: ## Generate Swagger docs dari kode Go
	@echo "Generating Swagger documentation..."
	swag init -g $(MAIN_FILE) -o docs
	@echo "Swagger docs generated di folder ./docs"


## ---------- ## mocks Commands ## ---------- #
# Generate mocks
.PHONY: mocks
mocks: ## Generate mocks untuk testing
	@echo "Generating mocks..."
	mockery --name URLRepository --dir repository/url --output mocks/repository/url --outpkg mocks --case underscore
	@echo "Mocks generated!"

# Generate all mocks in repository
.PHONY: mocks-all
mocks-all: ## Generate semua mocks di repository
	@echo "Generating all repository mocks..."
	@if not exist mocks\repository mkdir mocks\repository
	@echo "Generating mocks for repository/url..."
	@mockery --all --dir repository/url --output mocks/repository/url --outpkg mocks --case underscore
	@echo "All repository mocks generated!"

# Generate mocks for specific interface
.PHONY: mocks-url
mocks-url: ## Generate mock untuk URLRepository saja
	@echo "Generating URLRepository mock..."
	mockery --name URLRepository --dir repository/url --output mocks/repository/url --outpkg mocks --case underscore
	@echo "URLRepository mock generated!"

# Generate mocks for all layers
.PHONY: mocks-everything
mocks-everything: ## Generate mocks untuk semua layer (repository, service, external)
	@echo "Generating mocks for all layers..."
	@if not exist mocks mkdir mocks
	@echo "Generating repository mocks..."
	@mockery --all --dir repository/url --output mocks/repository/url --outpkg mocks --case underscore
	@echo "All layer mocks generated!"

## ---------- ## DB Commands ## ---------- #
# Database migration
.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@if [ -f "db/migrations/*.sql" ]; then \
		echo "Migration files found"; \
	else \
		echo "No migration files found in db/migrations/"; \
	fi

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@echo "Migration down not implemented yet"

.PHONY: migrate-status
migrate-status: ## Check migration status
	@echo "Migration status:"
	@echo "Migration status not implemented yet"


## ---------- ## workflow Commands ## ---------- #
# Development setup
.PHONY: dev-setup
dev-setup: deps mocks swagger ## Setup environment untuk development
	@echo "Development setup complete!"
	@echo "Run 'make run' to start the application"

# Super command - jalankan semua
.PHONY: all
all: clean deps mocks swagger test build ## Clean, setup, test, dan build semua
	@echo "All tasks completed successfully!"

# CI command - untuk continuous integration
.PHONY: ci
ci: deps mocks swagger test-coverage fmt vet ## Run all CI checks
	@echo "All CI checks passed!"

# Quick development workflow
.PHONY: dev
dev: deps mocks swagger test build ## Quick dev workflow: deps + mocks + test + build
	@echo "Development workflow complete!"
