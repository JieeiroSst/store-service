package voucher

import (
	"context"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
)

// ---- Driving port (inbound use cases) ----

type IssueVouchersInput struct {
	MerchantID     shared.MerchantID
	ProductSKU     string
	Denomination   shared.Money
	Quantity       int
	OrderID        *shared.OrderID
	ExpiresInDays  int
	IdempotencyKey string
}

type RedeemVoucherInput struct {
	VoucherID      shared.VoucherID
	PIN            string
	Amount         shared.Money
	IdempotencyKey string
}

type RedeemVoucherOutput struct {
	VoucherID      string       `json:"voucher_id"`
	Status         string       `json:"status"`
	RedeemedAmount shared.Money `json:"redeemed_amount"`
	ProviderTxnRef string       `json:"provider_txn_ref,omitempty"`
}

type ValidateVoucherInput struct {
	VoucherID shared.VoucherID
	PIN       string
}

type IssuedVoucher struct {
	Voucher      *voucher.Voucher
	PlaintextPIN string
}

type VoucherService interface {
	IssueVouchers(ctx context.Context, in IssueVouchersInput) ([]IssuedVoucher, error)
	ActivateVoucher(ctx context.Context, id shared.VoucherID, ownerType voucher.OwnerType, ownerID string) (*voucher.Voucher, error)
	RedeemVoucher(ctx context.Context, in RedeemVoucherInput) (*RedeemVoucherOutput, error)
	ValidateVoucher(ctx context.Context, in ValidateVoucherInput) (shared.ValidationResult, error)
	RevokeVoucher(ctx context.Context, id shared.VoucherID, reason string) error
	GetVoucher(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error)
	ListVouchers(ctx context.Context, ownerType voucher.OwnerType, ownerID string) ([]*voucher.Voucher, error)
	ExpireDueVouchers(ctx context.Context) (expired int, err error)
}

// ---- Driven ports (outbound) ----

// VoucherRepository is owned by the voucher context.
type VoucherRepository interface {
	Create(ctx context.Context, v *voucher.Voucher) error
	FindByID(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error)
	FindByIDForUpdate(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error)
	FindByCode(ctx context.Context, merchantID shared.MerchantID, code string) (*voucher.Voucher, error)
	ListByOwner(ctx context.Context, ownerType voucher.OwnerType, ownerID string) ([]*voucher.Voucher, error)
	ListDueForExpiry(ctx context.Context, now time.Time) ([]*voucher.Voucher, error)
	Save(ctx context.Context, v *voucher.Voucher) error
}

type MerchantProvider interface {
	Issue(ctx context.Context, ref shared.ProductRef, qty int) ([]shared.VoucherCode, error)
	Validate(ctx context.Context, code, pin string) (shared.ValidationResult, error)
	Redeem(ctx context.Context, code, pin string, amount shared.Money) (shared.RedeemResult, error)
	Type() shared.ProviderType
}

type ProviderRegistry interface {
	Resolve(providerType shared.ProviderType) (MerchantProvider, error)
}

type MerchantInfo struct {
	ID           shared.MerchantID
	ProviderType shared.ProviderType
	Active       bool
}

type MerchantLookup interface {
	GetMerchantInfo(ctx context.Context, id shared.MerchantID) (MerchantInfo, error)
}
