PKGS = ./...

run:
	EJABBERD_CONFIG=conf/ejabberd.yml go run ./cmd/ejabberd-service

build:
	go build $(PKGS)

vet:
	go vet $(PKGS)

test:
	go test -race $(PKGS)

docker:
	docker build -t ejabberd-service .
