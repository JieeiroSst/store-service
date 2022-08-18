gen:
	grpc_tools_node_protoc --js_out=import_style=commonjs,binary:proto/ --grpc_out=grpc_js:proto --proto_path=./protos/api ./proto/*.proto

install-grpc:
	npm_config_target_arch=x64 npm i grpc-tools 