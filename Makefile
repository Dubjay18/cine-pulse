.PHONY: build run test clean docker-build docker-run docker-stop dev logs migrate-up migrate-down migrate-status migrate-version migrate-reset

# Go build variables
BINARY_NAME=cine-pulse
MIGRATE_BINARY=migrate
BUILD_DIR=./bin

# Docker variables
DOCKER_IMAGE=cine-pulse
CONTAINER_NAME=cine-pulse-app

# Migration variables
DATA_PATH=./data

# Build the Go binary
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go

# Build migration tool
build-migrate:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(MIGRATE_BINARY) cmd/migrate/main.go

# Run locally (requires SQLite)
run: build
	DATA_PATH=./data $(BUILD_DIR)/$(BINARY_NAME)

# Run in scheduler mode
run-scheduler: build
	RUN_MODE=scheduler RUN_AT_STARTUP=true DATA_PATH=./data $(BUILD_DIR)/$(BINARY_NAME)

# Run once and exit
run-once: build
	RUN_MODE=once DATA_PATH=./data $(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run scheduler tests
test-scheduler:
	go test -v ./scheduler

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf ./data
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Migration commands
migrate-up: build-migrate
	$(BUILD_DIR)/$(MIGRATE_BINARY) -data $(DATA_PATH) -cmd up

migrate-down: build-migrate
	$(BUILD_DIR)/$(MIGRATE_BINARY) -data $(DATA_PATH) -cmd down

migrate-status: build-migrate
	$(BUILD_DIR)/$(MIGRATE_BINARY) -data $(DATA_PATH) -cmd status

migrate-version: build-migrate
	$(BUILD_DIR)/$(MIGRATE_BINARY) -data $(DATA_PATH) -cmd version

migrate-reset: build-migrate
	$(BUILD_DIR)/$(MIGRATE_BINARY) -data $(DATA_PATH) -cmd reset

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker run with compose
docker-run:
	docker-compose up --build

# Docker run in background
docker-run-bg:
	docker-compose up -d --build

# Docker run with development profile (includes SQLite browser)
dev:
	docker-compose --profile dev up --build

# Stop Docker containers
docker-stop:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Clean Docker resources
docker-clean:
	docker-compose down -v
	docker system prune -f

# Backup database
backup:
	@if [ ! -d "./backups" ]; then mkdir backups; fi
	docker cp $(CONTAINER_NAME):/data/cine_pulse.db ./backups/backup_$(shell date +%Y%m%d_%H%M%S).db
	@echo "Database backed up to ./backups/"

# Restore database (usage: make restore BACKUP=backup_file.db)
restore:
	@if [ -z "$(BACKUP)" ]; then echo "Usage: make restore BACKUP=backup_file.db"; exit 1; fi
	docker cp $(BACKUP) $(CONTAINER_NAME):/data/cine_pulse.db
	@echo "Database restored from $(BACKUP)"

# Development setup
setup:
	@if [ ! -f ".env" ]; then cp .env.example .env; echo "Created .env file. Please edit with your API keys."; fi
	go mod download

# Show database stats
stats:
	docker exec -it $(CONTAINER_NAME) sqlite3 /data/cine_pulse.db "SELECT COUNT(*) as total FROM content; SELECT type, COUNT(*) as count FROM content GROUP BY type;"

# Docker migration commands
docker-migrate-up:
	docker exec -it $(CONTAINER_NAME) /app/migrate -data /data -cmd up

docker-migrate-down:
	docker exec -it $(CONTAINER_NAME) /app/migrate -data /data -cmd down

docker-migrate-status:
	docker exec -it $(CONTAINER_NAME) /app/migrate -data /data -cmd status

docker-migrate-version:
	docker exec -it $(CONTAINER_NAME) /app/migrate -data /data -cmd version

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the Go binary"
	@echo "  build-migrate - Build migration tool"
	@echo "  run           - Run locally (requires SQLite)"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  migrate-up    - Run migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-status - Show migration status"
	@echo "  migrate-version - Show database version"
	@echo "  migrate-reset - Reset database"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-run-bg - Run in background"
	@echo "  dev           - Run with development profile (SQLite browser)"
	@echo "  docker-stop   - Stop Docker containers"
	@echo "  logs          - View container logs"
	@echo "  docker-clean  - Clean Docker resources"
	@echo "  backup        - Backup database"
	@echo "  restore       - Restore database (usage: make restore BACKUP=file.db)"
	@echo "  setup         - Development setup"
	@echo "  stats         - Show database statistics"
	@echo "  docker-migrate-* - Run migrations in Docker"
	@echo "  help          - Show this help"
