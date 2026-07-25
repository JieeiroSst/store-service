package grpc

import (
	"context"

	grpcapi "github.com/JIeeiroSst/lib-gateway/account-transaction-service/gateway/account-transaction-service"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/trace_id"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/port"
	"go.uber.org/zap"
)

type Handler struct {
	account     port.AccountUsecase
	transaction port.TransactionUsecase
	grpcapi.UnimplementedBankingServiceServer
}

func NewHandler(account port.AccountUsecase, transaction port.TransactionUsecase) *Handler {
	return &Handler{account: account, transaction: transaction}
}

// ─── Account operations ───────────────────────────────────────────────────────

func (h *Handler) CreateAccount(ctx context.Context, in *grpcapi.CreateAccountRequest) (*grpcapi.Account, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	account, err := h.account.CreateAccount(ctx, in.GetFirstName(), in.GetLastName())
	if err != nil {
		logger.WithContext(ctx).Error("CreateAccount", zap.Error(err))
		return nil, mapError(err)
	}
	return accountToPB(*account), nil
}

func (h *Handler) GetAccount(ctx context.Context, in *grpcapi.GetAccountRequest) (*grpcapi.Account, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	account, err := h.account.GetAccount(ctx, in.GetId())
	if err != nil {
		logger.WithContext(ctx).Error("GetAccount", zap.Error(err))
		return nil, mapError(err)
	}
	return accountToPB(*account), nil
}

func (h *Handler) ListAccounts(ctx context.Context, in *grpcapi.ListAccountsRequest) (*grpcapi.ListAccountsResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	accounts, nextPageToken, err := h.account.ListAccounts(ctx, in.GetPageSize(), in.GetPageToken())
	if err != nil {
		logger.WithContext(ctx).Error("ListAccounts", zap.Error(err))
		return nil, mapError(err)
	}
	return &grpcapi.ListAccountsResponse{
		Accounts:      accountsToPB(accounts),
		NextPageToken: nextPageToken,
	}, nil
}

// ─── Transaction operations ───────────────────────────────────────────────────

func (h *Handler) CreateDepositTransaction(ctx context.Context, in *grpcapi.CreateDepositTransactionRequest) (*grpcapi.Transaction, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	tx, err := h.transaction.CreateDeposit(ctx, in.GetAccountId(), in.GetAmount())
	if err != nil {
		logger.WithContext(ctx).Error("CreateDepositTransaction", zap.Error(err))
		return nil, mapError(err)
	}
	return transactionToPB(*tx), nil
}

func (h *Handler) CreateWithdrawalTransaction(ctx context.Context, in *grpcapi.CreateWithdrawalTransactionRequest) (*grpcapi.Transaction, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	tx, err := h.transaction.CreateWithdrawal(ctx, in.GetAccountId(), in.GetAmount())
	if err != nil {
		logger.WithContext(ctx).Error("CreateWithdrawalTransaction", zap.Error(err))
		return nil, mapError(err)
	}
	return transactionToPB(*tx), nil
}

func (h *Handler) CreateAccountToAccountTransaction(ctx context.Context, in *grpcapi.CreateAccountToAccountTransactionRequest) (*grpcapi.Transaction, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	tx, err := h.transaction.CreateTransfer(ctx, in.GetSenderId(), in.GetReceiverId(), in.GetAmount())
	if err != nil {
		logger.WithContext(ctx).Error("CreateAccountToAccountTransaction", zap.Error(err))
		return nil, mapError(err)
	}
	return transactionToPB(*tx), nil
}

func (h *Handler) CreatePaymentForServiceTransaction(ctx context.Context, in *grpcapi.CreatePaymentForServiceTransactionRequest) (*grpcapi.Transaction, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	tx, err := h.transaction.CreatePayment(ctx, in.GetAccountId(), in.GetServiceName(), in.GetAmount())
	if err != nil {
		logger.WithContext(ctx).Error("CreatePaymentForServiceTransaction", zap.Error(err))
		return nil, mapError(err)
	}
	return transactionToPB(*tx), nil
}

func (h *Handler) GetTransaction(ctx context.Context, in *grpcapi.GetTransactionRequest) (*grpcapi.Transaction, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	tx, err := h.transaction.GetTransaction(ctx, in.GetId())
	if err != nil {
		logger.WithContext(ctx).Error("GetTransaction", zap.Error(err))
		return nil, mapError(err)
	}
	return transactionToPB(*tx), nil
}

func (h *Handler) ListTransactions(ctx context.Context, in *grpcapi.ListTransactionsRequest) (*grpcapi.ListTransactionsResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	txs, nextPageToken, err := h.transaction.ListTransactions(ctx, in.GetPageSize(), in.GetPageToken(), in.GetType())
	if err != nil {
		logger.WithContext(ctx).Error("ListTransactions", zap.Error(err))
		return nil, mapError(err)
	}
	return &grpcapi.ListTransactionsResponse{
		Transactions:  transactionsToPB(txs),
		NextPageToken: nextPageToken,
	}, nil
}

func (h *Handler) GetAccountTransactions(ctx context.Context, in *grpcapi.GetAccountTransactionsRequest) (*grpcapi.ListTransactionsResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	txs, nextPageToken, err := h.transaction.GetAccountTransactions(ctx, in.GetAccountId(), in.GetPageSize(), in.GetPageToken())
	if err != nil {
		logger.WithContext(ctx).Error("GetAccountTransactions", zap.Error(err))
		return nil, mapError(err)
	}
	return &grpcapi.ListTransactionsResponse{
		Transactions:  transactionsToPB(txs),
		NextPageToken: nextPageToken,
	}, nil
}
