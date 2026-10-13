package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SardOpenAccount(ctx context.Context, input map[string]any) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: "/api/v1/accounts", Body: input}, "")
}

func (r *mutationResolver) SardBlockAccount(ctx context.Context, id string, reason *string) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/accounts/%v/block", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) SardUnblockAccount(ctx context.Context, id string, reason *string) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/accounts/%v/unblock", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) SardCancelAccount(ctx context.Context, id string, reason *string) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/accounts/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) SardSetCreditLimit(ctx context.Context, id string, input map[string]any) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/accounts/%v/credit-limit", id), Body: input}, "")
}

func (r *mutationResolver) SardAccountPayment(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/accounts/%v/payments", id), Body: input}, "")
}

func (r *mutationResolver) SardIssueCard(ctx context.Context, id string, input map[string]any) (*model.SardIssuedCard, error) {
	return restCall[*model.SardIssuedCard](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/accounts/%v/cards", id), Body: input}, "")
}

func (r *mutationResolver) SardActivateCard(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/activate", id), Body: input}, "")
}

func (r *mutationResolver) SardBlockCard(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/block", id), Body: input}, "")
}

func (r *mutationResolver) SardUnblockCard(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/unblock", id), Body: input}, "")
}

func (r *mutationResolver) SardCancelCard(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/cancel", id), Body: input}, "")
}

func (r *mutationResolver) SardReportCard(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/report", id), Body: input}, "")
}

func (r *mutationResolver) SardUpdateCardLimits(ctx context.Context, id string, perTransaction *int, daily *int) (*model.SardCard, error) {
	return restCall[*model.SardCard](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/cards/%v/limits", id), Body: map[string]any{"per_transaction": perTransaction, "daily": daily}}, "")
}

func (r *mutationResolver) SardUpdateCardControls(ctx context.Context, id string, pos *bool, contactless *bool, ecommerce *bool, atm *bool) (*model.SardCard, error) {
	return restCall[*model.SardCard](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/cards/%v/controls", id), Body: map[string]any{"pos": pos, "contactless": contactless, "ecommerce": ecommerce, "atm": atm}}, "")
}

func (r *mutationResolver) SardSetCardPin(ctx context.Context, id string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/cards/%v/pin", id), Body: input}, "")
}

func (r *mutationResolver) SardAuthorize(ctx context.Context, input map[string]any) (*model.SardAuthorization, error) {
	return restCall[*model.SardAuthorization](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: "/api/v1/authorizations", Body: input}, "")
}

func (r *mutationResolver) SardConfirmAuthorization(ctx context.Context, id string, input map[string]any) (*model.SardAuthorization, error) {
	return restCall[*model.SardAuthorization](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/authorizations/%v/confirm", id), Body: input}, "")
}

func (r *mutationResolver) SardCancelAuthorization(ctx context.Context, id string, input map[string]any) (*model.SardAuthorization, error) {
	return restCall[*model.SardAuthorization](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/authorizations/%v/cancel", id), Body: input}, "")
}

func (r *queryResolver) SardPrograms(ctx context.Context) ([]*model.SardProgram, error) {
	return restCall[[]*model.SardProgram](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: "/api/v1/programs"}, "")
}

func (r *queryResolver) SardAccount(ctx context.Context, id string) (*model.SardAccount, error) {
	return restCall[*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/accounts/%v", id)}, "")
}

func (r *queryResolver) SardCustomerAccounts(ctx context.Context, customerID string) ([]*model.SardAccount, error) {
	return restCall[[]*model.SardAccount](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/customers/%v/accounts", customerID)}, "")
}

func (r *queryResolver) SardAccountTransactions(ctx context.Context, id string, query map[string]any) ([]*model.SardTransaction, error) {
	return restCall[[]*model.SardTransaction](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/accounts/%v/transactions", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) SardAccountCards(ctx context.Context, id string) ([]*model.SardCard, error) {
	return restCall[[]*model.SardCard](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/accounts/%v/cards", id)}, "")
}

func (r *queryResolver) SardCard(ctx context.Context, id string) (*model.SardCard, error) {
	return restCall[*model.SardCard](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/cards/%v", id)}, "")
}

func (r *queryResolver) SardCardAuthorizations(ctx context.Context, id string, query map[string]any) ([]*model.SardAuthorization, error) {
	return restCall[[]*model.SardAuthorization](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/cards/%v/authorizations", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) SardAuthorization(ctx context.Context, id string) (*model.SardAuthorization, error) {
	return restCall[*model.SardAuthorization](ctx, r.Proxy, "sard-card-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/authorizations/%v", id)}, "")
}
