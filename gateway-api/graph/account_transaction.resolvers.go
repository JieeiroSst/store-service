package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
)

func (r *mutationResolver) AccountTxCreateAccount(ctx context.Context, input model.AccountTxCreateAccountRequestInput) (*model.AccountTxAccount, error) {
	return grpcInvoke[model.AccountTxAccount](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/CreateAccount", "banking.CreateAccountRequest", "banking.Account", input)
}

func (r *mutationResolver) AccountTxCreateDepositTransaction(ctx context.Context, input model.AccountTxCreateDepositTransactionRequestInput) (*model.AccountTxTransaction, error) {
	return grpcInvoke[model.AccountTxTransaction](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/CreateDepositTransaction", "banking.CreateDepositTransactionRequest", "banking.Transaction", input)
}

func (r *mutationResolver) AccountTxCreateWithdrawalTransaction(ctx context.Context, input model.AccountTxCreateWithdrawalTransactionRequestInput) (*model.AccountTxTransaction, error) {
	return grpcInvoke[model.AccountTxTransaction](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/CreateWithdrawalTransaction", "banking.CreateWithdrawalTransactionRequest", "banking.Transaction", input)
}

func (r *mutationResolver) AccountTxCreateAccountToAccountTransaction(ctx context.Context, input model.AccountTxCreateAccountToAccountTransactionRequestInput) (*model.AccountTxTransaction, error) {
	return grpcInvoke[model.AccountTxTransaction](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/CreateAccountToAccountTransaction", "banking.CreateAccountToAccountTransactionRequest", "banking.Transaction", input)
}

func (r *mutationResolver) AccountTxCreatePaymentForServiceTransaction(ctx context.Context, input model.AccountTxCreatePaymentForServiceTransactionRequestInput) (*model.AccountTxTransaction, error) {
	return grpcInvoke[model.AccountTxTransaction](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/CreatePaymentForServiceTransaction", "banking.CreatePaymentForServiceTransactionRequest", "banking.Transaction", input)
}

func (r *queryResolver) AccountTxGetAccount(ctx context.Context, input model.AccountTxGetAccountRequestInput) (*model.AccountTxAccount, error) {
	return grpcInvoke[model.AccountTxAccount](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/GetAccount", "banking.GetAccountRequest", "banking.Account", input)
}

func (r *queryResolver) AccountTxListAccounts(ctx context.Context, input model.AccountTxListAccountsRequestInput) (*model.AccountTxListAccountsResponse, error) {
	return grpcInvoke[model.AccountTxListAccountsResponse](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/ListAccounts", "banking.ListAccountsRequest", "banking.ListAccountsResponse", input)
}

func (r *queryResolver) AccountTxGetTransaction(ctx context.Context, input model.AccountTxGetTransactionRequestInput) (*model.AccountTxTransaction, error) {
	return grpcInvoke[model.AccountTxTransaction](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/GetTransaction", "banking.GetTransactionRequest", "banking.Transaction", input)
}

func (r *queryResolver) AccountTxListTransactions(ctx context.Context, input model.AccountTxListTransactionsRequestInput) (*model.AccountTxListTransactionsResponse, error) {
	return grpcInvoke[model.AccountTxListTransactionsResponse](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/ListTransactions", "banking.ListTransactionsRequest", "banking.ListTransactionsResponse", input)
}

func (r *queryResolver) AccountTxGetAccountTransactions(ctx context.Context, input model.AccountTxGetAccountTransactionsRequestInput) (*model.AccountTxListTransactionsResponse, error) {
	return grpcInvoke[model.AccountTxListTransactionsResponse](ctx, r.Proxy, proxy.AccountTransactionService, "/banking.BankingService/GetAccountTransactions", "banking.GetAccountTransactionsRequest", "banking.ListTransactionsResponse", input)
}
