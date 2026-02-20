.PHONY: run build test tidy docker-up docker-down lint

APP_NAME = qr-service
BUILD_DIR = ./bin
MAIN = ./cmd/main.go

run:
	go run $(MAIN)

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

test:
	go test -v -race ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

.DEFAULT_GOAL := run
