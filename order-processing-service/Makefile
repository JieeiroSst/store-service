.PHONY: worker starter cron build clean

# Start the Temporal Worker
worker:
	go run ./cmd/worker/main.go

# Start a new Order Processing Workflow
starter:
	go run ./cmd/starter/main.go

# Start the Cron Cleanup Workflow
cron:
	go run ./cmd/cron/main.go

# Build all binaries
build:
	go build -o bin/worker ./cmd/worker
	go build -o bin/starter ./cmd/starter
	go build -o bin/cron ./cmd/cron

# Run tests
test:
	go test ./... -v -count=1

# Clean build artifacts
clean:
	rm -rf bin/

# Start local Temporal server (requires temporal CLI)
temporal-server:
	temporal server start-dev --ui-port 8233

# Tidy dependencies
tidy:
	go mod tidy
