.PHONY: help build run test clean docker-up docker-down migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/chatbot-server ./cmd/api

run: ## Run the application locally
	go run ./cmd/api/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-rebuild: ## Rebuild and restart Docker services
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

migrate: ## Run database migrations
	docker exec -i chatbot_mysql mysql -u root -ppassword chatbot_db < migrations/001_init_schema.sql

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run || go vet ./...

dev: docker-up ## Start development environment
	@echo "Waiting for MySQL to be ready..."
	@sleep 10
	@make migrate
	@echo ""
	@echo "✅ Development environment is ready!"
	@echo "🌐 Server: http://localhost:8080"
	@echo "🗄️  MySQL: localhost:3306"
	@echo "📊 Health check: http://localhost:8080/health"
	@echo ""

stop: docker-down ## Stop development environment

check-health: ## Check if server is running
	@curl -s http://localhost:8080/health | jq . || echo "Server is not running"
