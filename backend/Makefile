.PHONY: help build run dev clean test migrate migrate-up migrate-down seed-reelstrips seed-assets db-create db-drop db-reset tidy rtp-check rtp-tuning

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := slotmachine
SERVER_BIN := ./bin/server
SEED_SCRIPT := ./scripts/seed_reelstrips.sh
RTP_SCRIPT := ./scripts/rtp.sh
TUNING_RTP_SCRIPT := ./scripts/rtp_tuning.sh

# Load .env file if exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Default DB values if not in .env
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= slotmachine
DB_SSL_MODE ?= disable

# Database migration variables
MIGRATIONS_DIR := ./migrations
DATABASE_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the server binary
build:
	@echo "🔨 Building server..."
	@mkdir -p bin
	@go build -o $(SERVER_BIN) ./cmd/server
	@echo "✅ Build complete: $(SERVER_BIN)"

## run: Run the server (production mode)
run: build
	@echo "🚀 Starting server..."
	@$(SERVER_BIN)

## dev: Run server with hot reload using Air
dev:
	@echo "🔥 Starting development server with Air..."
	@if ! command -v air > /dev/null; then \
		echo "⚠️  Air not found. Installing..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@air

## clean: Clean build artifacts and tmp files
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin tmp build-errors.log
	@echo "✅ Clean complete"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "📊 Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## migrate: Run all database migrations using golang-migrate
migrate:
	@echo "🗄️  Running database migrations..."
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "✅ Migrations complete"

## migrate-up: Run all migrations (alias for migrate)
migrate-up: migrate

## migrate-down: Rollback last migration
migrate-down:
	@echo "⚠️  Rolling back last migration..."
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "✅ Rollback complete"

## migrate-down-all: Rollback all migrations
migrate-down-all:
	@echo "⚠️  Rolling back ALL migrations..."
	@read -p "Are you sure? This will drop all tables! [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all; \
		echo "✅ All migrations rolled back"; \
	else \
		echo "❌ Cancelled"; \
	fi

## migrate-version: Show current migration version
migrate-version:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

## migrate-force: Force migration version (use with caution)
migrate-force:
	@echo "⚠️  Forcing migration version to $(VERSION)..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "✅ Version forced to $(VERSION)"

## migrate-create: Create a new migration file (use: make migrate-create NAME=create_users)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required. Usage: make migrate-create NAME=create_users"; \
		exit 1; \
	fi
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "✅ Migration files created"

## seed-reelstrips: Seed reel strips (args: MODE=both CREATE_CONFIG=true)
seed-reelstrips:
	@echo "🎰 Seeding reel strips..."
	@chmod +x $(SEED_SCRIPT)
	@$(SEED_SCRIPT) $(MODE) $(CREATE_CONFIG)

## seed-reelstrips-base: Seed base game reel strips with default config
seed-reelstrips-base:
	@$(MAKE) seed-reelstrips MODE=base_game CREATE_CONFIG=true

## seed-reelstrips-free: Seed free spins reel strips with default config
seed-reelstrips-free:
	@$(MAKE) seed-reelstrips MODE=free_spins CREATE_CONFIG=true

## seed-reelstrips-both: Seed both game modes with default configs
seed-reelstrips-both:
	@$(MAKE) seed-reelstrips MODE=both CREATE_CONFIG=true

## seed-assets: Seed default game assets (Bamboo theme, Mahjong Ways game)
seed-assets:
	@echo "🎨 Seeding game assets..."
	@chmod +x ./scripts/seed_assets.sh
	@./scripts/seed_assets.sh

db-create:
	echo "📦 Creating database..."; \
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME)";

db-drop:
	echo "🗑️  Dropping database..."; \
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME)" 2>/dev/null || true;


## db-reset: Drop and recreate database, run migrations and seed
db-reset:
	@echo "⚠️  This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "🗑️  Dropping database..."; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME)" 2>/dev/null || true; \
		echo "📦 Creating database..."; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME)"; \
		echo "🗄️  Running migrations..."; \
		$(MAKE) migrate; \
		echo "🎨 Seeding game assets..."; \
		$(MAKE) seed-assets; \
		echo "🎰 Seeding reel strips..."; \
		$(MAKE) seed-reelstrips-both; \
		echo "✅ Database reset complete"; \
	else \
		echo "❌ Cancelled"; \
	fi

## db-status: Check database connection and show table counts
db-status:
	@echo "📊 Database Status:"
	@echo "===================="
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "\
		SELECT \
			tablename as table_name, \
			schemaname as schema \
		FROM pg_tables \
		WHERE schemaname = 'public' \
		ORDER BY tablename;" 2>/dev/null || echo "❌ Database not accessible"
	@echo ""
	@echo "Table Counts:"
	@echo "-------------"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -t -c "\
		SELECT 'players: ' || COUNT(*) FROM players;" 2>/dev/null || echo "players: (table not found)"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -t -c "\
		SELECT 'reel_strips: ' || COUNT(*) || ' (active: ' || COUNT(*) FILTER (WHERE is_active = true) || ')' FROM reel_strips;" 2>/dev/null || echo "reel_strips: (table not found)"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -t -c "\
		SELECT 'game_sessions: ' || COUNT(*) FROM game_sessions;" 2>/dev/null || echo "game_sessions: (table not found)"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -t -c "\
		SELECT 'spins: ' || COUNT(*) FROM spins;" 2>/dev/null || echo "spins: (table not found)"
	@echo ""
	@echo "Migration Version:"
	@echo "------------------"
	@$(MAKE) migrate-version 2>/dev/null || echo "No migrations applied yet"

## rtp-check: Run RTP simulation
rtp-check:
	@echo "🎲 Running RTP simulation..."
	@chmod +x $(RTP_SCRIPT)
	@$(RTP_SCRIPT)

## rtp-tuning: Run RTP simulation
rtp-tuning:
	@echo "🎲 Tunning RTP..."
	@chmod +x $(TUNING_RTP_SCRIPT)
	@$(TUNING_RTP_SCRIPT)

## tidy: Tidy go modules
tidy:
	@echo "📦 Tidying go modules..."
	@go mod tidy
	@echo "✅ Tidy complete"

## fmt: Format Go code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "🔍 Running linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "⚠️  golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@golangci-lint run ./...
	@echo "✅ Lint complete"

## wire: Generate Wire dependency injection code
wire:
	@echo "⚡ Generating Wire code..."
	@if ! command -v wire > /dev/null; then \
		echo "⚠️  Wire not found. Installing..."; \
		go install github.com/google/wire/cmd/wire@latest; \
	fi
	@cd cmd/server && wire
	@cd scripts/seed_reelstrips && wire
	@cd scripts/seed_assets && wire
	@echo "✅ Wire generation complete"

## wire-check: Check Wire configuration without generating
wire-check:
	@echo "🔍 Checking Wire configuration..."
	@if ! command -v wire > /dev/null; then \
		echo "⚠️  Wire not found. Installing..."; \
		go install github.com/google/wire/cmd/wire@latest; \
	fi
	@cd cmd/server && wire check
	@cd scripts/seed_reelstrips && wire check
	@cd scripts/seed_assets && wire check
	@echo "✅ Wire check complete"

## install-tools: Install development tools
install-tools:
	@echo "🔧 Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/google/wire/cmd/wire@latest
	@echo "✅ Tools installed (air, golangci-lint, golang-migrate, wire)"

## setup: Initial setup (install tools, migrate, seed)
setup: install-tools migrate seed-assets seed-reelstrips-both
	@echo "✅ Setup complete! Run 'make dev' to start development server"

## prod-deploy: Build and prepare for production deployment
prod-deploy: clean test build
	@echo "📦 Production build ready: $(SERVER_BIN)"
	@echo "✅ Deploy with: scp $(SERVER_BIN) user@server:/path/to/app"

## stats: Show project statistics
stats:
	@echo "📊 Project Statistics:"
	@echo "===================="
	@echo "Go files: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Lines of code: $$(find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "Test files: $$(find . -name '*_test.go' | wc -l)"
	@echo "Migrations: $$(ls -1 migrations/*.sql 2>/dev/null | wc -l)"
