package auth

import "context"

type Client struct {
	IP        string
	UserAgent string
}

type clientKey struct{}

func WithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func ClientFromContext(ctx context.Context) Client {
	c, _ := ctx.Value(clientKey{}).(Client)
	return c
}
