package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SardVerifyPaymentAuthentication(ctx context.Context, id string, otp string) (*model.SardChallenge, error) {
	return restCall[*model.SardChallenge](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/me/payment-authentications/%v/verify", id), Body: map[string]any{"otp": otp}}, "")
}

func (r *mutationResolver) SardResendPaymentAuthentication(ctx context.Context, id string) (*model.SardChallenge, error) {
	return restCall[*model.SardChallenge](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/me/payment-authentications/%v/resend", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) SardDeclinePaymentAuthentication(ctx context.Context, id string) (*model.SardChallenge, error) {
	return restCall[*model.SardChallenge](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/me/payment-authentications/%v/decline", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) SardInitiatePaymentAuthentication(ctx context.Context, input map[string]any) (*model.SardChallenge, error) {
	return restCall[*model.SardChallenge](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "POST", Path: "/api/v1/payment-authentications", Body: input}, "")
}

func (r *queryResolver) MySardPaymentAuthentications(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "GET", Path: "/api/v1/me/payment-authentications"}, "")
}

func (r *queryResolver) SardPaymentAuthentication(ctx context.Context, id string) (*model.SardChallenge, error) {
	return restCall[*model.SardChallenge](ctx, r.Proxy, "sard-auth-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payment-authentications/%v", id)}, "")
}
