package http

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/car-rental-service/internal/auth"
	"github.com/JIeeiroSst/car-rental-service/internal/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeUsers map[string]userservice.Identity

func (f fakeUsers) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	if id, ok := f[token]; ok {
		return id, nil
	}
	return userservice.Identity{}, userservice.ErrUnauthenticated
}

// asCaller runs fn with the claims the gRPC interceptor would attach.
func asCaller(t *testing.T, a *auth.Authenticator, token string, fn func(ctx context.Context)) {
	t.Helper()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
	_, err := a.UnaryInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/any"},
		func(ctx context.Context, _ any) (any, error) { fn(ctx); return nil, nil })
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
}

func TestActingUser(t *testing.T) {
	users := fakeUsers{
		"alice": {UserID: "1", Role: "user"},
		"staff": {UserID: "2", Role: "operator"},
	}
	a, _ := auth.New(auth.Config{}, users, nil)
	h := &Handler{auth: a}

	asCaller(t, a, "alice", func(ctx context.Context) {
		if id, err := h.actingUser(ctx, ""); err != nil || id != "1" {
			t.Errorf("empty user_id: got %q, %v; want caller 1", id, err)
		}
		if id, err := h.actingUser(ctx, "1"); err != nil || id != "1" {
			t.Errorf("own user_id: got %q, %v", id, err)
		}
		if _, err := h.actingUser(ctx, "2"); status.Code(err) != codes.PermissionDenied {
			t.Errorf("customer acting on another user: got %v, want PermissionDenied", err)
		}
		if !h.canAccess(ctx, 1) || h.canAccess(ctx, 2) {
			t.Error("customer should only access own records")
		}
	})
	asCaller(t, a, "staff", func(ctx context.Context) {
		if id, err := h.actingUser(ctx, "1"); err != nil || id != "1" {
			t.Errorf("staff acting on user 1: got %q, %v", id, err)
		}
		if !h.canAccess(ctx, 1) {
			t.Error("staff should access any record")
		}
		if got := staffID(ctx); got != "2" {
			t.Errorf("staffID = %q, want 2", got)
		}
	})
	if _, err := h.actingUser(context.Background(), "1"); status.Code(err) != codes.Unauthenticated {
		t.Errorf("anonymous: got %v, want Unauthenticated", err)
	}
}
