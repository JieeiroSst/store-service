package auth

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/car-rental-service/internal/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// fakeUsers maps token → role; "down" simulates user-service being unreachable.
type fakeUsers map[string]string

func (f fakeUsers) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	if token == "down" {
		return userservice.Identity{}, userservice.ErrUpstream
	}
	role, ok := f[token]
	if !ok {
		return userservice.Identity{}, userservice.ErrUnauthenticated
	}
	return userservice.Identity{UserID: "7", Username: "u", Role: role}, nil
}

func call(a *Authenticator, method, tok string) error {
	ctx := context.Background()
	if tok != "" {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer "+tok))
	}
	_, err := a.UnaryInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method},
		func(context.Context, any) (any, error) { return nil, nil })
	return err
}

func TestInterceptor(t *testing.T) {
	users := fakeUsers{"user": "user", "operator": "operator", "admin": "admin", "root": "super_admin"}
	a, _ := New(Config{}, users, map[string]Level{"/pub": Public, "/staff": Staff, "/admin": Admin})
	tests := []struct {
		name, method, tok string
		want              codes.Code
	}{
		{"public anonymous", "/pub", "", codes.OK},
		{"unlisted needs login", "/other", "", codes.Unauthenticated},
		{"unlisted with token", "/other", "user", codes.OK},
		{"customer on staff", "/staff", "user", codes.PermissionDenied},
		{"operator on staff", "/staff", "operator", codes.OK},
		{"operator on admin", "/admin", "operator", codes.PermissionDenied},
		{"admin on admin", "/admin", "admin", codes.OK},
		{"super_admin on admin", "/admin", "root", codes.OK},
		{"rejected by user-service", "/pub", "forged", codes.Unauthenticated},
		{"user-service down", "/other", "down", codes.Unavailable},
	}
	for _, tt := range tests {
		if got := status.Code(call(a, tt.method, tt.tok)); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
