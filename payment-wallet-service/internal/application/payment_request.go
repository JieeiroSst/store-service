package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/google/uuid"
)

type paymentRequestService struct {
	wallets      port.WalletRepository
	requests     port.PaymentRequestRepository
	transactions port.TransactionUsecase
}

func NewPaymentRequestService(wallets port.WalletRepository, requests port.PaymentRequestRepository, transactions port.TransactionUsecase) port.PaymentRequestUsecase {
	return &paymentRequestService{wallets: wallets, requests: requests, transactions: transactions}
}

func (s *paymentRequestService) CreatePaymentRequest(ctx context.Context, requesterWalletID string, payerWalletID *string, amount int64, description string, expiresIn time.Duration) (*model.PaymentRequest, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	requester, err := s.wallets.GetByID(ctx, requesterWalletID)
	if err != nil {
		return nil, fmt.Errorf("get requester wallet: %w", err)
	}

	request := &model.PaymentRequest{
		PaymentRequestID:  uuid.NewString(),
		RequesterWalletID: requesterWalletID,
		PayerWalletID:     payerWalletID,
		Amount:            amount,
		Currency:          requester.Currency,
		Status:            model.PaymentRequestPending,
		Description:       description,
	}
	if expiresIn > 0 {
		expiresAt := time.Now().Add(expiresIn)
		request.ExpiresAt = &expiresAt
	}
	if err := s.requests.Create(ctx, request); err != nil {
		return nil, fmt.Errorf("create payment request: %w", err)
	}
	return request, nil
}

func (s *paymentRequestService) GetPaymentRequest(ctx context.Context, id string) (*model.PaymentRequest, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *paymentRequestService) ListPaymentRequests(ctx context.Context, walletID string) ([]model.PaymentRequest, error) {
	return s.requests.ListByWallet(ctx, walletID)
}

func (s *paymentRequestService) Pay(ctx context.Context, id, payerWalletID string) (*model.Transfer, error) {
	request, err := s.requests.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get payment request: %w", err)
	}
	if request.Status != model.PaymentRequestPending {
		return nil, port.ErrPaymentRequestNotPending
	}
	if request.Expired(time.Now()) {
		_ = s.requests.UpdateStatus(ctx, id, model.PaymentRequestExpired, nil)
		return nil, port.ErrPaymentRequestExpired
	}
	if request.PayerWalletID != nil && *request.PayerWalletID != payerWalletID {
		return nil, port.ErrPaymentRequestPayerMismatch
	}

	transfer, err := s.transactions.Transfer(ctx, payerWalletID, request.RequesterWalletID, request.Amount, "", fmt.Sprintf("payment request %s", id))
	if err != nil {
		return nil, fmt.Errorf("pay: %w", err)
	}
	if err := s.requests.UpdateStatus(ctx, id, model.PaymentRequestPaid, &transfer.TransferID); err != nil {
		return nil, fmt.Errorf("mark payment request paid: %w", err)
	}
	return transfer, nil
}

func (s *paymentRequestService) Cancel(ctx context.Context, id string) error {
	request, err := s.requests.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get payment request: %w", err)
	}
	if request.Status != model.PaymentRequestPending {
		return port.ErrPaymentRequestNotPending
	}
	return s.requests.UpdateStatus(ctx, id, model.PaymentRequestCancelled, nil)
}
