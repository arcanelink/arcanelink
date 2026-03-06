.PHONY: help build run stop clean test proto migrate migrate-up migrate-down migrate-status db-shell db-reset test-file-upload file-storage-info file-storage-clean web-client web-client-build web-client-test

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
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/008_create_file_storage.up.sql

migrate-down: ## Run database migrations down
	@echo "Rolling back migrations..."
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/008_create_file_storage.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/007_create_presence.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/006_create_message_queue.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/005_create_room_events.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/004_create_room_members.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/003_create_rooms.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/002_create_direct_messages.down.sql
	@docker-compose exec postgres psql -U mbot -d mbot -f /docker-entrypoint-initdb.d/001_create_users.down.sql

migrate-status: ## Check migration status
	@echo "Checking database tables..."
	@docker-compose exec postgres psql -U mbot -d mbot -c "\dt"

db-shell: ## Open PostgreSQL shell
	@docker-compose exec postgres psql -U mbot -d mbot

db-reset: migrate-down migrate-up ## Reset database (down then up)
	@echo "Database reset complete"

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

test-file-upload: ## Test file upload functionality
	@echo "Testing file upload..."
	@chmod +x scripts/test_file_upload.sh
	@./scripts/test_file_upload.sh

file-storage-info: ## Show file storage information
	@echo "File storage information:"
	@echo "Storage path: ./data/files (or FILE_STORAGE_PATH env var)"
	@echo ""
	@echo "Files in storage:"
	@docker-compose exec api-gateway find /data/files -type f 2>/dev/null || echo "No files found or service not running"
	@echo ""
	@echo "Database records:"
	@docker-compose exec postgres psql -U mbot -d mbot -c "SELECT file_id, filename, content_type, file_size, uploader, created_at FROM file_storage ORDER BY created_at DESC LIMIT 10;" 2>/dev/null || echo "Database not accessible"

file-storage-clean: ## Clean up file storage (WARNING: deletes all files)
	@echo "WARNING: This will delete all uploaded files!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Cleaning file storage..."; \
		docker-compose exec api-gateway rm -rf /data/files/* 2>/dev/null || echo "Could not clean files"; \
		docker-compose exec postgres psql -U mbot -d mbot -c "DELETE FROM file_storage;" 2>/dev/null || echo "Could not clean database"; \
		echo "File storage cleaned"; \
	else \
		echo "Cancelled"; \
	fi

web-client: ## Start web client development server
	@echo "Starting web client..."
	@cd web-client && npm install && npm run dev

web-client-build: ## Build web client for production
	@echo "Building web client..."
	@cd web-client && npm install && npm run build

web-client-test: ## Run web client tests
	@echo "Running web client tests..."
	@cd web-client && npm test
