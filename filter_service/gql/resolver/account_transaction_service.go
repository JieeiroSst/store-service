package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AcctTxQuery() generated.AcctTxQueryResolver { return &acctTxQueryResolver{r} }

type acctTxQueryResolver struct{ *Resolver }

func (r *acctTxQueryResolver) GetAccount(ctx context.Context, obj *model.AcctTxQuery, id string) (*model.AcctTxAccount, error) {
	return r.Clients.AccountTransactionService.GetAccount(ctx, id)
}

func (r *acctTxQueryResolver) ListAccounts(ctx context.Context, obj *model.AcctTxQuery, pageSize *int, pageToken *string) (*model.AcctTxListAccountsResponse, error) {
	return r.Clients.AccountTransactionService.ListAccounts(ctx, pageSize, pageToken)
}

func (r *acctTxQueryResolver) GetTransaction(ctx context.Context, obj *model.AcctTxQuery, id string) (*model.AcctTxTransaction, error) {
	return r.Clients.AccountTransactionService.GetTransaction(ctx, id)
}

func (r *acctTxQueryResolver) ListTransactions(ctx context.Context, obj *model.AcctTxQuery, pageSize *int, pageToken *string, typeArg *string) (*model.AcctTxListTransactionsResponse, error) {
	return r.Clients.AccountTransactionService.ListTransactions(ctx, pageSize, pageToken, typeArg)
}

func (r *acctTxQueryResolver) GetAccountTransactions(ctx context.Context, obj *model.AcctTxQuery, accountID string, pageSize *int, pageToken *string) (*model.AcctTxListTransactionsResponse, error) {
	return r.Clients.AccountTransactionService.GetAccountTransactions(ctx, accountID, pageSize, pageToken)
}
