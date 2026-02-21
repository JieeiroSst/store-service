.PHONY: run build test migrate lint

APP_NAME = wallet-service
BUILD_DIR = ./bin

run:
	go run ./cmd/main.go

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

test:
	go test ./... -v -race

migrate:
	psql -U postgres -d wallet_service -f migrations/001_init.sql

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down
