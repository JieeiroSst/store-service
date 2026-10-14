GATEWAY_GRAPH ?= ../gateway-api/graph

sync:
	go run ./cmd/catalogsync -src $(GATEWAY_GRAPH) -dst catalog/schema

sync-check: sync
	git diff --exit-code -- catalog/schema
	test -z "$$(git status --porcelain -- catalog/schema)"

upstream:
	rm -rf .upstream && mkdir -p .upstream/gateway-api
	cp -R $(GATEWAY_GRAPH) .upstream/gateway-api/graph

generate:
	go run github.com/99designs/gqlgen generate

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

run:
	go run .
