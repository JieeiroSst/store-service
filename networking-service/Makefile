PKGS = ./...

run:
	go run ./cmd

build:
	go build $(PKGS)

vet:
	go vet $(PKGS)

test:
	go test -race $(PKGS)

docker:
	docker build -t networking-service .
