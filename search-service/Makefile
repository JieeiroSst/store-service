PKGS = ./...
ES_TAG = 7.17.29

run:
	go run ./cmd

build:
	go build $(PKGS)

vet:
	go vet $(PKGS)

test:
	go test -race $(PKGS)

docker:
	docker build -t search-service .

up:
	TAG=$(ES_TAG) docker-compose up

down:
	TAG=$(ES_TAG) docker-compose down
