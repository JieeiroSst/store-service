run:
	go run cmd/server.go
gen:
	go run github.com/99designs/gqlgen@v0.17.27 generate

gen_auth:
	protoc --go_out=. --go-grpc_out=. ./adapter/authorize_service/apis.proto