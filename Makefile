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

##@ Testing Commands

.PHONY: test
test: ## Run all tests
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests with short flag
	$(GOTEST) -short -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race: ## Run tests with race detection
	$(GOTEST) -race -v ./...

.PHONY: bench
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

##@ Database Commands

.PHONY: db-up
db-up: ## Run database migrations
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

.PHONY: db-down
db-down: ## Rollback database migrations
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

.PHONY: db-reset
db-reset: ## Reset database (down then up)
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

.PHONY: db-force
db-force: ## Force database version (usage: make db-force VERSION=1)
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(VERSION)

.PHONY: db-version
db-version: ## Show current database version
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

.PHONY: db-create-migration
db-create-migration: ## Create new migration file (usage: make db-create-migration NAME=add_users_table)
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)

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

##@ GraphQL Commands

.PHONY: gql-gen
gql-gen: ## Generate GraphQL code
	go run github.com/99designs/gqlgen generate

.PHONY: gql-init
gql-init: ## Initialize GraphQL (run once)
	go run github.com/99designs/gqlgen init

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