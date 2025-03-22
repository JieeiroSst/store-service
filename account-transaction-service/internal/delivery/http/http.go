package http

import (
	"context"

	grpc "github.com/JIeeiroSst/lib-gateway/account-transaction-service/gateway/account-transaction-service"
	"github.com/Jieeirosst/account-transaction-service/internal/usecase"
)

type Handler struct {
	usecase *usecase.Usecase
	grpc.UnimplementedBankingServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// Account operations
func (h *Handler) CreateAccount(ctx context.Context, in *grpc.CreateAccountRequest) (*grpc.Account, error) {
	return nil, nil
}
func (h *Handler) GetAccount(ctx context.Context, in *grpc.GetAccountRequest) (*grpc.Account, error) {
	return nil, nil
}

func (h *Handler) ListAccounts(ctx context.Context, in *grpc.ListAccountsRequest) (*grpc.ListAccountsResponse, error) {
	return nil, nil
}

// Transaction operations
func (h *Handler) CreateDepositTransaction(ctx context.Context, in *grpc.CreateDepositTransactionRequest) (*grpc.Transaction, error) {
	return nil, nil
}
func (h *Handler) CreateWithdrawalTransaction(ctx context.Context, in *grpc.CreateWithdrawalTransactionRequest) (*grpc.Transaction, error) {
	return nil, nil
}

func (h *Handler) CreateAccountToAccountTransaction(ctx context.Context, in *grpc.CreateAccountToAccountTransactionRequest) (*grpc.Transaction, error) {
	return nil, nil
}
func (h *Handler) CreatePaymentForServiceTransaction(ctx context.Context, in *grpc.CreatePaymentForServiceTransactionRequest) (*grpc.Transaction, error) {
	return nil, nil
}
func (h *Handler) GetTransaction(ctx context.Context, in *grpc.GetTransactionRequest) (*grpc.Transaction, error) {
	return nil, nil
}
func (h *Handler) ListTransactions(ctx context.Context, in *grpc.ListTransactionsRequest) (*grpc.ListTransactionsResponse, error) {
	return nil, nil
}
func (h *Handler) GetAccountTransactions(ctx context.Context, in *grpc.GetAccountTransactionsRequest) (*grpc.ListTransactionsResponse, error) {
	return nil, nil
}
