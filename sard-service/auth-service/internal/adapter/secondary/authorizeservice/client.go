package authorizeservice

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var Module = fx.Options(
	fx.Provide(New),
)

func New(lc fx.Lifecycle, cfg *config.Config) (port.OTPProvider, error) {
	conn, err := grpc.NewClient(cfg.AuthorizeService.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("authorize-service client: %w", err)
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return conn.Close() }})
	return NewClient(authorizeGrpc.NewAuthorizeServiceClient(conn), cfg.AuthorizeService.Timeout), nil
}

type Client struct {
	rpc     authorizeGrpc.AuthorizeServiceClient
	timeout time.Duration
}

func NewClient(rpc authorizeGrpc.AuthorizeServiceClient, timeout time.Duration) *Client {
	return &Client{rpc: rpc, timeout: timeout}
}

func (c *Client) Issue(ctx context.Context, username string) (string, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	res, err := c.rpc.CreateOTP(ctx, &authorizeGrpc.CreateOTPRequest{Username: username})
	if err != nil {
		return "", time.Time{}, mapError(err)
	}
	if len(res.Otp) != domain.OTPLength {
		return "", time.Time{}, fmt.Errorf("%w: authorize-service returned an otp of length %d", domain.ErrUnavailable, len(res.Otp))
	}
	if res.ExpiresAt <= 0 {
		return "", time.Time{}, fmt.Errorf("%w: authorize-service returned no otp expiry", domain.ErrUnavailable)
	}
	return res.Otp, time.Unix(res.ExpiresAt, 0).UTC(), nil
}

func (c *Client) Verify(ctx context.Context, username, code string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	res, err := c.rpc.AuthorizeOTP(ctx, &authorizeGrpc.AuthorizeOTPRequest{Username: username, Otp: code})
	if err != nil {
		return false, mapError(err)
	}
	return res.Valid, nil
}

func mapError(err error) error {
	switch status.Code(err) {
	case codes.ResourceExhausted:
		return domain.ErrOTPLimit
	case codes.InvalidArgument:
		return domain.Invalid("authorize-service rejected the otp request: %s", status.Convert(err).Message())
	}
	return fmt.Errorf("%w: authorize-service: %s", domain.ErrUnavailable, strconv.Quote(status.Convert(err).Message()))
}
