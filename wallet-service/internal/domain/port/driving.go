package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
)

type WalletService interface {
	Create(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error)
	Get(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error)
	GetByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Wallet, error)
	Balance(ctx context.Context, walletID uuid.UUID) (*model.Balance, error)
	Credit(ctx context.Context, cmd CreditCommand) (*model.Transaction, error)
	Freeze(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error)
	Unfreeze(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error)
	Close(ctx context.Context, walletID uuid.UUID, userID uuid.UUID) (*model.Wallet, error)
	ChangeTier(ctx context.Context, walletID uuid.UUID, tier model.Tier) (*model.Wallet, error)
	Limits(ctx context.Context, walletID uuid.UUID) (*model.LimitUsage, error)
	SetPin(ctx context.Context, cmd SetPinCommand) error
}

type FundingService interface {
	TopUp(ctx context.Context, cmd TopUpCommand) (*model.Transaction, error)
	Withdraw(ctx context.Context, cmd WithdrawCommand) (*model.Transaction, error)
}

type TransferService interface {
	Transfer(ctx context.Context, cmd TransferCommand) (*model.Transaction, error)
}

type PaymentService interface {
	Pay(ctx context.Context, cmd PayCommand) (*model.Transaction, error)
	Refund(ctx context.Context, cmd RefundCommand) (*model.Transaction, error)
}

type CardPaymentService interface {
	Authorize(ctx context.Context, cmd AuthorizeCommand) (*model.Transaction, error)
	Capture(ctx context.Context, txID uuid.UUID) (*model.Transaction, error)
	Void(ctx context.Context, txID uuid.UUID) (*model.Transaction, error)
}

type SettlementService interface {
	CreateBatch(ctx context.Context, merchantID uuid.UUID) (*model.SettlementBatch, error)
	Clear(ctx context.Context, batchID uuid.UUID) ([]*model.ClearingRecord, error)
	Settle(ctx context.Context, batchID uuid.UUID) error
}

type HistoryService interface {
	Transaction(ctx context.Context, txID uuid.UUID) (*model.Transaction, error)
	Transactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, int64, error)
	Ledger(ctx context.Context, walletID uuid.UUID, from, to *time.Time, page, pageSize int) ([]*model.LedgerEntry, int64, error)
	Statement(ctx context.Context, walletID uuid.UUID, from, to time.Time) (*model.Statement, error)
}

type CardService interface {
	Issue(ctx context.Context, cmd IssueCardCommand) (*model.Card, error)
	Get(ctx context.Context, cardID uuid.UUID) (*model.Card, error)
	ListByWallet(ctx context.Context, walletID uuid.UUID) ([]*model.Card, error)
	Block(ctx context.Context, cardID uuid.UUID) (*model.Card, error)
	Unblock(ctx context.Context, cardID uuid.UUID) (*model.Card, error)
}

type BankAccountService interface {
	Link(ctx context.Context, cmd LinkBankAccountCommand) (*model.BankAccount, error)
	List(ctx context.Context, userID uuid.UUID) ([]*model.BankAccount, error)
	Remove(ctx context.Context, userID, accountID uuid.UUID) error
	SetDefault(ctx context.Context, userID, accountID uuid.UUID) (*model.BankAccount, error)
}

type PaymentRequestService interface {
	Create(ctx context.Context, cmd CreatePaymentRequestCommand) (*model.PaymentRequest, error)
	Get(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error)
	List(ctx context.Context, filter model.PaymentRequestFilter) ([]*model.PaymentRequest, int64, error)
	Pay(ctx context.Context, cmd PayPaymentRequestCommand) (*model.PaymentRequest, error)
	Decline(ctx context.Context, id, userID uuid.UUID) (*model.PaymentRequest, error)
	Cancel(ctx context.Context, id, userID uuid.UUID) (*model.PaymentRequest, error)
}

type CreditCommand struct {
	WalletID       uuid.UUID
	Amount         decimal.Decimal
	Description    string
	IdempotencyKey string
}

type SetPinCommand struct {
	WalletID   uuid.UUID
	UserID     uuid.UUID
	CurrentPin string
	NewPin     string
}

type TopUpCommand struct {
	IdempotencyKey string
	UserID         uuid.UUID
	WalletID       uuid.UUID
	BankAccountID  uuid.UUID
	Amount         decimal.Decimal
	Description    string
}

type WithdrawCommand struct {
	IdempotencyKey string
	UserID         uuid.UUID
	WalletID       uuid.UUID
	BankAccountID  uuid.UUID
	Amount         decimal.Decimal
	Pin            string
	Description    string
}

type TransferCommand struct {
	IdempotencyKey string
	UserID         uuid.UUID
	FromWalletID   uuid.UUID
	ToWalletID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         decimal.Decimal
	Pin            string
	Description    string
}

type PayCommand struct {
	IdempotencyKey string
	UserID         uuid.UUID
	WalletID       uuid.UUID
	MerchantID     uuid.UUID
	Amount         decimal.Decimal
	Pin            string
	Description    string
	Metadata       map[string]string
}

type RefundCommand struct {
	IdempotencyKey string
	TransactionID  uuid.UUID
	Amount         decimal.Decimal
	Reason         string
}

type AuthorizeCommand struct {
	IdempotencyKey string
	CardID         uuid.UUID
	MerchantID     uuid.UUID
	Amount         decimal.Decimal
	Currency       model.Currency
	Description    string
	Metadata       map[string]string
}

type IssueCardCommand struct {
	WalletID     uuid.UUID
	IssuerBankID uuid.UUID
	Network      model.CardNetwork
	HolderName   string
}

type LinkBankAccountCommand struct {
	UserID        uuid.UUID
	BankID        uuid.UUID
	AccountNumber string
	AccountName   string
}

type CreatePaymentRequestCommand struct {
	RequesterUserID   uuid.UUID
	RequesterWalletID uuid.UUID
	PayerUserID       uuid.UUID
	Amount            decimal.Decimal
	Note              string
}

type PayPaymentRequestCommand struct {
	RequestID uuid.UUID
	UserID    uuid.UUID
	WalletID  uuid.UUID
	Pin       string
}
