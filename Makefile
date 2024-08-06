include .env

# Variables
ENV_FILE = .env
DOCKER_COMPOSE_FILE = docker-compose.yml
PROJECT_NAME = gophermart

# Targets
.PHONY: all build up down logs clean develop lint

# Default target
all: build up

# Build Docker images
build:
	@echo "Building Docker images..."
	docker compose --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --project-name $(PROJECT_NAME) build

# Start containers in the background
up:
	@echo "Starting Docker containers..."
	docker compose --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --project-name $(PROJECT_NAME) up --build --detach

# Stop and remove containers
down:
	@echo "Stopping and removing Docker containers..."
	docker compose --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --project-name $(PROJECT_NAME) down

# View logs
logs:
	@echo "Viewing logs for Docker containers..."
	docker compose --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --project-name $(PROJECT_NAME) logs --follow

# Clean up unused data
clean:
	@echo "Cleaning up unused Docker data..."
	docker system prune --all --force

# Start development mode with file watching
develop:
	@echo "Starting development mode with file watching..."
	docker compose --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --project-name $(PROJECT_NAME) up --watch

# Lint the code
lint:
	@echo "Running linter..."
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v --fix

# Create a new migration
db_migration_new:
	@echo "Creating a new migration..."
	@read -p "Enter migration name: " name; \
	docker run --rm -v $(shell pwd)/internal/storage/postgres/migrations:/migrations \
		migrate/migrate create -ext sql -dir /migrations -seq -digits 5 $${name}

# Apply migrations (up)
db_migrate_up:
	@echo "Applying migrations..."
	docker compose run --rm migrate -path /migrations -database ${DATABASE_URI} up

# Rollback migrations (down)
db_migrate_down:
	@echo "Rolling back migrations..."
	docker compose run --rm migrate -path /migrations -database ${DATABASE_URI} down 1

# Running tests
test:
	@echo "Running tests..."
	docker run --rm -v $(shell pwd):/app -w /app golang:1.22.5 go test ./...

# Generate mocks
mocks:
	@echo "Generating mocks..."
	docker run --rm -v $(shell pwd):/app -w /app vektra/mockery --all