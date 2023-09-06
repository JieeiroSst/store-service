Run protoc để gen

service.pb.go

    protoc --go_out=. --go_opt plugins=grpc --go_opt paths=source_relative .\proto\service.proto


service.pb.gw.go

    protoc --grpc-gateway_out=. --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative .\proto\service.proto


Run server và proxy
    
    go run .\main.go
    go run .\proxy\proxy.go


test 

    http://localhost:8081/v1/echo
    http://localhost:8081/v1/sum?num1=1&num2=2