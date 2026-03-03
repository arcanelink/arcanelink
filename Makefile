.PHONY: help build run stop clean test proto migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building services..."
	@docker-compose build

run: ## Start all services
	@echo "Starting services..."
	@docker-compose up -d
	@echo "Services started. API Gateway available at http://localhost:8080"

stop: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

clean: ## Stop services and remove volumes
	@echo "Cleaning up..."
	@docker-compose down -v
	@rm -rf postgres_data redis_data

logs: ## Show logs from all services
	@docker-compose logs -f

logs-api: ## Show API Gateway logs
	@docker-compose logs -f api-gateway

logs-auth: ## Show Auth Service logs
	@docker-compose logs -f auth

logs-message: ## Show Message Service logs
	@docker-compose logs -f message

logs-room: ## Show Room Service logs
	@docker-compose logs -f room

logs-federation: ## Show Federation Service logs
	@docker-compose logs -f federation

logs-presence: ## Show Presence Service logs
	@docker-compose logs -f presence

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	@go tool cover -html=coverage.out

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/*.proto

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/001_create_users.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/002_create_direct_messages.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/003_create_rooms.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/004_create_room_members.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/005_create_room_events.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/006_create_message_queue.up.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/007_create_presence.up.sql

ps: ## Show running services
	@docker-compose ps

restart: stop run ## Restart all services

dev: ## Start services in development mode with live reload
	@echo "Starting in development mode..."
	@docker-compose up

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

build-local: ## Build binaries locally (requires Go installed)
	@echo "Building binaries..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./cmd/api-gateway
	@go build -o bin/auth ./cmd/auth
	@go build -o bin/message ./cmd/message
	@go build -o bin/room ./cmd/room
	@go build -o bin/federation ./cmd/federation
	@go build -o bin/presence ./cmd/presence
	@echo "Binaries built in ./bin/"
