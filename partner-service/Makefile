grpc:
	protoc --go_out=. --go_opt plugins=grpc --go_opt paths=source_relative .\proto\*.proto
api:
	protoc --grpc-gateway_out=. --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative .\proto\*.proto
