package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *queryResolver) PayerTransaction(ctx context.Context, transactionID int) (*model.PayerTransactionResponse, error) {
	return restCall[*model.PayerTransactionResponse](ctx, r.Proxy, "payer-service", rest.Request{Method: "GET", Path: "/api/v1/transaction", Query: restQuery("transaction_id", transactionID)}, "data")
}
