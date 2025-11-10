# Project configuration
APP_NAME=my-go-next-todo
MAIN_PATH=./cmd/server
BIN_DIR=./bin
MIGRATIONS_PATH=./migrations

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Build flags
BUILD_FLAGS=-ldflags="-w -s"
DEV_BUILD_FLAGS=-race

# Database configuration (will use DATABASE_URL from .env if available)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=todoapp
DB_URL=$(shell if [ -n "$$DATABASE_URL" ]; then echo "$$DATABASE_URL"; else echo "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=require"; fi)

# Default target
.DEFAULT_GOAL := help

##@ Development Commands

.PHONY: dev
dev: ## Start development server with hot reload
	air

.PHONY: run
run: ## Run application without hot reload
	$(GORUN) $(MAIN_PATH)

.PHONY: build
build: ## Build production binary
	mkdir -p $(BIN_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

.PHONY: build-dev
build-dev: ## Build development binary with race detection
	mkdir -p $(BIN_DIR)
	$(GOBUILD) $(DEV_BUILD_FLAGS) -o $(BIN_DIR)/$(APP_NAME)-dev $(MAIN_PATH)

.PHONY: setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env 2>/dev/null || echo "Please create .env file manually"; \
	fi
	@echo "Environment ready! Run 'make dev' to start development server"

.PHONY: deps-all
deps-all: mod-download deps ## Install all dependencies
	@echo "All dependencies installed"

##@ Testing Commands

.PHONY: test-unit
test-unit: ## Run unit tests (fast)
	$(GOTEST) -v ./internal/...

.PHONY: test-integration
test-integration: ## Run integration tests (slow, needs server)
	$(GOTEST) -v ./tests/...

.PHONY: test-all
test-all: test-unit test-integration ## Run all tests
.PHONY: bench
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

##@ Database Commands

.PHONY: db-up
db-up: ## Run database migrations
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" up

.PHONY: db-down
db-down: ## Rollback database migrations
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" down

.PHONY: db-reset
db-reset: ## Reset database (down then up)
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" down
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" up

.PHONY: db-force
db-force: ## Force database version (usage: make db-force VERSION=1)
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION is required. Usage: make db-force VERSION=1"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" force $(VERSION)

.PHONY: db-version
db-version: ## Show current database version
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" version

.PHONY: db-create-migration
db-create-migration: ## Create new migration file (usage: make db-create-migration NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Usage: make db-create-migration NAME=migration_name"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)

.PHONY: db-status
db-status: ## Show migration status and database info
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@echo "Database Migration Status:"
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" version
	@echo "\nAvailable migrations:"
	@ls -la $(MIGRATIONS_PATH)/*.sql 2>/dev/null || echo "No migration files found"

.PHONY: db-tables
db-tables: ## List all tables in database
	@if [ ! -f .env ]; then echo "Error: .env file not found"; exit 1; fi
	@set -a; source .env; set +a; psql "$$DATABASE_URL" -c "\dt"
	
.PHONY: db-drop
db-drop: ## Drop all tables (DANGEROUS - use with caution)
	@echo "WARNING: This will drop all tables in the database!"
	@echo "Are you sure? Type 'yes' to continue:"
	@read confirmation; if [ "$$confirmation" != "yes" ]; then echo "Aborted."; exit 1; fi
	@if [ ! -f .env ]; then echo "Error: .env file not found. Please create .env file with DATABASE_URL"; exit 1; fi
	@set -a; source .env; set +a; migrate -path $(MIGRATIONS_PATH) -database "$$DATABASE_URL" drop

##@ GraphQL Commands

.PHONY: gql-generate
gql-generate: ## Generate GraphQL code
	gqlgen generate

.PHONY: gql-init
gql-init: ## Initialize GraphQL (run once)
	go run github.com/99designs/gqlgen init

.PHONY: gql-validate
gql-validate: ## Validate GraphQL schema
	go run github.com/99designs/gqlgen validate

.PHONY: gql-initialize
gql-initialize: ## Initialize GraphQL (alternative)
	go run github.com/99designs/gqlgen init

##@ Code Quality Commands

.PHONY: fmt
fmt: ## Format Go code
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix

.PHONY: mod-tidy
mod-tidy: ## Tidy Go modules
	$(GOMOD) tidy

.PHONY: mod-download
mod-download: ## Download Go modules
	$(GOMOD) download

.PHONY: mod-verify
mod-verify: ## Verify Go modules
	$(GOMOD) verify

##@ Frontend Commands (Next.js)

.PHONY: frontend-install
frontend-install: ## Install frontend dependencies
	cd frontend && npm install

.PHONY: frontend-dev
frontend-dev: ## Start frontend development server
	cd frontend && npm run dev

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	cd frontend && npm run build

.PHONY: frontend-test
frontend-test: ## Run frontend tests
	cd frontend && npm test

##@ Docker Commands (for later use)

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 -e DB_URL="$(DB_URL)" $(APP_NAME)

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	docker-compose down

##@ Utility Commands

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	rm -rf $(BIN_DIR)
	rm -rf tmp
	rm -f coverage.out coverage.html
	$(GOCMD) clean

.PHONY: deps
deps: ## Install development dependencies
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: full-test
full-test: test-race test-cover ## Run comprehensive tests

.PHONY: pre-commit
pre-commit: fmt vet lint test-short ## Run pre-commit checks

##@ Information Commands

.PHONY: info
info: ## Show project information
	@echo "Project: $(APP_NAME)"
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Database URL: $(DB_URL)"
	@echo "Main path: $(MAIN_PATH)"
	@echo "Binary directory: $(BIN_DIR)"

.PHONY: help
help: ## Display available commands
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)