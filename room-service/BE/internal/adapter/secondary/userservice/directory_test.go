package userservice

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	pb "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeClient struct {
	pb.UserServiceClient // unimplemented methods panic, which is what we want in a test
	validate             func(*pb.ValidateRequest) (*pb.ValidateResponse, error)
	calls                int
}

func (f *fakeClient) ValidateSession(_ context.Context, in *pb.ValidateRequest, _ ...grpc.CallOption) (*pb.ValidateResponse, error) {
	f.calls++
	return f.validate(in)
}

func (f *fakeClient) Login(context.Context, *pb.LoginRequest, ...grpc.CallOption) (*pb.LoginResponse, error) {
	return nil, errors.New("password entered incorrectly") // how user-service reports it
}

func jwt(payload string) string {
	return "h." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".s"
}

func newDir(c *fakeClient, ttl time.Duration) *Directory {
	cfg := &config.Config{}
	cfg.UserService.Timeout = time.Second
	cfg.UserService.AuthCacheTTL = ttl
	return NewDirectory(c, cfg)
}

func TestValidate(t *testing.T) {
	ok := func(*pb.ValidateRequest) (*pb.ValidateResponse, error) {
		return &pb.ValidateResponse{Valid: true, UserId: "42"}, nil
	}
	token := jwt(`{"sub":"42","username":"alice","role":"user"}`)

	c := &fakeClient{validate: ok}
	d := newDir(c, time.Minute)
	u, err := d.Validate(context.Background(), token)
	if err != nil || u.ID != 42 || u.Username != "alice" {
		t.Fatalf("Validate = %+v, %v", u, err)
	}
	if _, err := d.Validate(context.Background(), token); err != nil || c.calls != 1 {
		t.Fatalf("second call should be cached: err=%v calls=%d", err, c.calls)
	}

	// no caching when the TTL is zero
	c = &fakeClient{validate: ok}
	d = newDir(c, 0)
	_, _ = d.Validate(context.Background(), token)
	_, _ = d.Validate(context.Background(), token)
	if c.calls != 2 {
		t.Fatalf("calls = %d, want 2", c.calls)
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]struct {
		token string
		resp  *pb.ValidateResponse
		err   error
		want  error
	}{
		"revoked":         {jwt(`{"username":"a"}`), &pb.ValidateResponse{Valid: false}, nil, port.ErrUnauthenticated},
		"no username":     {jwt(`{"sub":"1"}`), &pb.ValidateResponse{Valid: true, UserId: "1"}, nil, port.ErrUnauthenticated},
		"not a jwt":       {"garbage", &pb.ValidateResponse{Valid: true, UserId: "1"}, nil, port.ErrUnauthenticated},
		"bad user id":     {jwt(`{"username":"a"}`), &pb.ValidateResponse{Valid: true, UserId: "x"}, nil, port.ErrUnauthenticated},
		"upstream down":   {jwt(`{"username":"a"}`), nil, status.Error(codes.Unavailable, "down"), port.ErrUnavailable},
		"upstream reject": {jwt(`{"username":"a"}`), nil, status.Error(codes.Unauthenticated, "no"), port.ErrUnauthenticated},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &fakeClient{validate: func(*pb.ValidateRequest) (*pb.ValidateResponse, error) { return tc.resp, tc.err }}
			if _, err := newDir(c, time.Minute).Validate(context.Background(), tc.token); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestLoginMapsFailureToInvalidLogin(t *testing.T) {
	_, err := newDir(&fakeClient{}, 0).Login(context.Background(), "a", "b")
	if !errors.Is(err, port.ErrInvalidLogin) {
		t.Fatalf("err = %v", err)
	}
}
