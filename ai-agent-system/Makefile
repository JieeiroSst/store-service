PKGS = ./...

.PHONY: run build vet fmt-check test docker db-up

run:
	go run ./cmd/server

build:
	go build $(PKGS)

vet:
	go vet $(PKGS)

fmt-check:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt needed"; exit 1; }

test:
	go test -race $(PKGS)

docker:
	docker build -t ai-agent-system .

db-up:
	docker compose up -d mysql
