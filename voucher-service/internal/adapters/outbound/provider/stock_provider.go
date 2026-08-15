package provider

import (
	"context"

	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type StockProvider struct {
	stock inventoryapp.StockClaimer
}

func NewStockProvider(stock inventoryapp.StockClaimer) voucherapp.MerchantProvider {
	return &StockProvider{stock: stock}
}

func (p *StockProvider) Type() shared.ProviderType { return shared.ProviderTypeStock }

func (p *StockProvider) Issue(ctx context.Context, ref shared.ProductRef, qty int) ([]shared.VoucherCode, error) {
	codes := make([]shared.VoucherCode, 0, qty)
	for i := 0; i < qty; i++ {
		code, pin, err := p.stock.ClaimCode(ctx, ref.MerchantID, ref.SKU, shared.NewVoucherID())
		if err != nil {
			return nil, err
		}
		codes = append(codes, shared.VoucherCode{Code: code, PIN: pin})
	}
	return codes, nil
}

func (p *StockProvider) Validate(ctx context.Context, code, pin string) (shared.ValidationResult, error) {
	return shared.ValidationResult{Valid: true}, nil
}

func (p *StockProvider) Redeem(ctx context.Context, code, pin string, amount shared.Money) (shared.RedeemResult, error) {
	return shared.RedeemResult{Success: true, RedeemedAmount: amount}, nil
}
