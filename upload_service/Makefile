run:
	go run cmd/main.go

mod:
	go mod tidy
gen:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/apis.proto

dep: ## Install required packages
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.35.2
	@go install github.com/rakyll/statik@v0.1.7
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
	@go install github.com/gogo/protobuf/protoc-gen-gogo@v1.3.1
	@go install github.com/favadi/protoc-go-inject-tag@v1.3.0
	@go install github.com/mwitkow/go-proto-validators/protoc-gen-govalidators@v0.3.2
	@go install github.com/vektra/mockery/cmd/mockery@v1.1.2
	@go install github.com/golang/mock/mockgen@v1.6.0

mockgen:
	mockgen --build_flags=--mod=mod -package usecase github.com/JIeeiroSst/upload-service/internal/repository Uploads > ./internal/usecase/mock_upload_test.go
.PHONY: run gen