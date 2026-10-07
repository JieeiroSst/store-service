package grpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	otpadapter "github.com/JieeiroSst/authorize-service/internal/adapter/secondary/otp"
	"github.com/JieeiroSst/authorize-service/internal/application"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type memCache struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *memCache) GetInterface(context.Context, string, interface{}) (interface{}, error) {
	return nil, nil
}
func (c *memCache) Set(context.Context, string, interface{}, time.Duration) error { return nil }
func (c *memCache) Delete(context.Context, string) error                          { return nil }
func (c *memCache) GetInt(_ context.Context, k string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[k], nil
}
func (c *memCache) SetInt(_ context.Context, k string, v int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
	return nil
}

func TestOTPOverGRPC(t *testing.T) {
	uc := application.NewOTPService(otpadapter.NewOTPAdapter("secretkey"), &memCache{m: map[string]int{}})
	lis := bufconn.Listen(1 << 20)
	srv := grpclib.NewServer()
	authorizeGrpc.RegisterAuthorizeServiceServer(srv, NewHandler(nil, uc))
	go srv.Serve(lis)
	defer srv.Stop()
	conn, err := grpclib.NewClient("passthrough:///bufnet",
		grpclib.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := authorizeGrpc.NewAuthorizeServiceClient(conn)
	ctx := context.Background()

	created, err := client.CreateOTP(ctx, &authorizeGrpc.CreateOTPRequest{Username: "quan_luu01"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	ttl := time.Until(time.Unix(created.ExpiresAt, 0))
	if len(created.Otp) != 6 || ttl <= 29*time.Second || ttl > 61*time.Second {
		t.Fatalf("otp %q ttl %s", created.Otp, ttl)
	}
	ok, err := client.AuthorizeOTP(ctx, &authorizeGrpc.AuthorizeOTPRequest{Username: "quan_luu01", Otp: created.Otp})
	if err != nil || !ok.Valid {
		t.Fatalf("real code rejected: %v %+v", err, ok)
	}
	wrong := "000000"
	if created.Otp == wrong {
		wrong = "111111"
	}
	if bad, err := client.AuthorizeOTP(ctx, &authorizeGrpc.AuthorizeOTPRequest{Username: "quan_luu01", Otp: wrong}); err != nil || bad.Valid {
		t.Fatalf("wrong code accepted: %v %+v", err, bad)
	}
	if other, _ := client.AuthorizeOTP(ctx, &authorizeGrpc.AuthorizeOTPRequest{Username: "someone_else", Otp: created.Otp}); other.Valid {
		t.Fatal("code accepted for another user")
	}
	for i := 0; i < 4; i++ {
		if _, err := client.CreateOTP(ctx, &authorizeGrpc.CreateOTPRequest{Username: "quan_luu01"}); err != nil {
			t.Fatalf("create %d: %v", i+2, err)
		}
	}
	if _, err := client.CreateOTP(ctx, &authorizeGrpc.CreateOTPRequest{Username: "quan_luu01"}); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("6th otp in 24h: %v", err)
	}
}
