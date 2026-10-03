PKGS = ./...

run:
	go run ./cmd

gen:
	go run ./tools/gen
	go tool gqlgen generate

build:
	go build $(PKGS)

vet:
	go vet $(PKGS)

test:
	go test -race $(PKGS)

# Fails when the committed schema, clients or resolvers are out of date with spec/.
check-gen: gen
	@test -z "$$(git status --porcelain -- gql adapter)" || (git status --porcelain -- gql adapter; echo "generated code is stale: run make gen and commit"; exit 1)

gen_auth:
	protoc --go_out=. --go-grpc_out=. ./adapter/authorize_service/apis.proto
