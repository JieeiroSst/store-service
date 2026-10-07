package authorizeservice

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type fakeAuthorize struct {
	authorizeGrpc.UnimplementedAuthorizeServiceServer
	issued int
}

func (f *fakeAuthorize) CreateOTP(_ context.Context, in *authorizeGrpc.CreateOTPRequest) (*authorizeGrpc.CreateOTPResponse, error) {
	switch in.Username {
	case "limited":
		return nil, status.Error(codes.ResourceExhausted, "OTP creation limit exceeded")
	case "down":
		return nil, status.Error(codes.Unavailable, "redis down")
	}
	f.issued++
	return &authorizeGrpc.CreateOTPResponse{Otp: "482913", ExpiresAt: 1_800_000_060}, nil
}

func (f *fakeAuthorize) AuthorizeOTP(_ context.Context, in *authorizeGrpc.AuthorizeOTPRequest) (*authorizeGrpc.AuthorizeOTPResponse, error) {
	return &authorizeGrpc.AuthorizeOTPResponse{Valid: in.Username == "anguyen" && in.Otp == "482913"}, nil
}

func TestClient(t *testing.T) {
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	fake := &fakeAuthorize{}
	authorizeGrpc.RegisterAuthorizeServiceServer(srv, fake)
	go srv.Serve(lis)
	defer srv.Stop()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	c := NewClient(authorizeGrpc.NewAuthorizeServiceClient(conn), time.Second)
	ctx := context.Background()

	code, exp, err := c.Issue(ctx, "anguyen")
	if err != nil || code != "482913" || exp.Unix() != 1_800_000_060 || fake.issued != 1 {
		t.Fatalf("issue: %q %v %v", code, exp, err)
	}
	if ok, err := c.Verify(ctx, "anguyen", "482913"); err != nil || !ok {
		t.Fatalf("verify: %v %v", ok, err)
	}
	if ok, err := c.Verify(ctx, "anguyen", "000000"); err != nil || ok {
		t.Fatalf("wrong code: %v %v", ok, err)
	}
	if _, _, err := c.Issue(ctx, "limited"); !errors.Is(err, domain.ErrOTPLimit) {
		t.Fatalf("limit: %v", err)
	}
	if _, _, err := c.Issue(ctx, "down"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("down: %v", err)
	}
}
