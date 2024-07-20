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
