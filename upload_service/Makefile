run:
	go run cmd/main.go

mod:
	go mod tidy
gen:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/apis.proto

image:
	docker build -t upload_service:${version}  .
image_tag:
	docker tag upload_service:${version}  1760151/upload_service:${version} 
image_push:
	docker push 1760151/upload_service:${version}

.PHONY: run gen