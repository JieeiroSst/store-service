package grpcadapter

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
)

func TestHandlerImplementsEveryRPC(t *testing.T) {
	server := reflect.TypeOf((*userServiceGrpc.UserServiceServer)(nil)).Elem()
	handler := reflect.TypeOf(&Handler{})
	for i := 0; i < server.NumMethod(); i++ {
		name := server.Method(i).Name
		if name == "mustEmbedUnimplementedUserServiceServer" {
			continue
		}
		m, ok := handler.MethodByName(name)
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		file, _ := runtime.FuncForPC(m.Func.Pointer()).FileLine(m.Func.Pointer())
		if !strings.HasSuffix(file, "grpcadapter/handler.go") {
			t.Errorf("%s falls back to the Unimplemented server (%s)", name, file)
		}
	}
}
