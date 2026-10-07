package grpc

import (
	"context"
	"testing"
	"time"

	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JieeiroSst/authorize-service/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeOTP struct {
	limited bool
}

func (f fakeOTP) CreateOtpByUser(_ context.Context, username string) (string, time.Time, error) {
	if f.limited {
		return "", time.Time{}, common.ErrOTPLimit
	}
	return "123456", time.Unix(1_800_000_060, 0), nil
}

func (f fakeOTP) Authorize(_ context.Context, code, username string) error {
	if code != "123456" {
		return common.ErrOTPFailed
	}
	return nil
}

func TestCreateOTP(t *testing.T) {
	h := NewHandler(nil, fakeOTP{})
	res, err := h.CreateOTP(context.Background(), &authorizeGrpc.CreateOTPRequest{Username: "alice"})
	if err != nil || res.Otp != "123456" || res.ExpiresAt != 1_800_000_060 {
		t.Fatalf("%v %+v", err, res)
	}
	if _, err := h.CreateOTP(context.Background(), &authorizeGrpc.CreateOTPRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty username: %v", err)
	}
	if _, err := NewHandler(nil, fakeOTP{limited: true}).CreateOTP(context.Background(), &authorizeGrpc.CreateOTPRequest{Username: "alice"}); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("limit: %v", err)
	}
}

func TestAuthorizeOTP(t *testing.T) {
	h := NewHandler(nil, fakeOTP{})
	ok, err := h.AuthorizeOTP(context.Background(), &authorizeGrpc.AuthorizeOTPRequest{Username: "alice", Otp: "123456"})
	if err != nil || !ok.Valid {
		t.Fatalf("valid: %v %+v", err, ok)
	}
	bad, err := h.AuthorizeOTP(context.Background(), &authorizeGrpc.AuthorizeOTPRequest{Username: "alice", Otp: "000000"})
	if err != nil || bad.Valid {
		t.Fatalf("invalid code is a normal answer, not an error: %v %+v", err, bad)
	}
	if _, err := h.AuthorizeOTP(context.Background(), &authorizeGrpc.AuthorizeOTPRequest{Otp: "1"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("missing username: %v", err)
	}
}
