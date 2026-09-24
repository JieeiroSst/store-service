package auth

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func token(t *testing.T, secret, role string, method jwt.SigningMethod) string {
	t.Helper()
	tk := jwt.NewWithClaims(method, jwt.MapClaims{"sub": "7", "username": "u", "role": role, "exp": time.Now().Add(time.Hour).Unix()})
	var key interface{} = []byte(secret)
	if method == jwt.SigningMethodNone {
		key = jwt.UnsafeAllowNoneSignatureType
	}
	s, err := tk.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
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
	a, _ := New(Config{SecretKey: "s"}, map[string]Level{"/pub": Public, "/staff": Staff, "/admin": Admin})
	tests := []struct {
		name, method, tok string
		want              codes.Code
	}{
		{"public anonymous", "/pub", "", codes.OK},
		{"unlisted needs login", "/other", "", codes.Unauthenticated},
		{"unlisted with token", "/other", token(t, "s", "user", jwt.SigningMethodHS256), codes.OK},
		{"customer on staff", "/staff", token(t, "s", "user", jwt.SigningMethodHS256), codes.PermissionDenied},
		{"staff on staff", "/staff", token(t, "s", "staff", jwt.SigningMethodHS256), codes.OK},
		{"staff on admin", "/admin", token(t, "s", "staff", jwt.SigningMethodHS256), codes.PermissionDenied},
		{"admin on admin", "/admin", token(t, "s", "admin", jwt.SigningMethodHS256), codes.OK},
		{"wrong secret", "/pub", token(t, "x", "admin", jwt.SigningMethodHS256), codes.Unauthenticated},
		{"alg none", "/pub", token(t, "s", "admin", jwt.SigningMethodNone), codes.Unauthenticated},
		{"HS512 rejected", "/pub", token(t, "s", "admin", jwt.SigningMethodHS512), codes.Unauthenticated},
	}
	for _, tt := range tests {
		if got := status.Code(call(a, tt.method, tt.tok)); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
