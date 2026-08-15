package provider

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"strings"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type SelfProvider struct{}

func NewSelfProvider() voucherapp.MerchantProvider {
	return &SelfProvider{}
}

func (p *SelfProvider) Type() shared.ProviderType { return shared.ProviderTypeSelf }

func (p *SelfProvider) Issue(ctx context.Context, ref shared.ProductRef, qty int) ([]shared.VoucherCode, error) {
	codes := make([]shared.VoucherCode, 0, qty)
	for i := 0; i < qty; i++ {
		code, err := randomCode(10)
		if err != nil {
			return nil, err
		}
		pin, err := randomCode(6)
		if err != nil {
			return nil, err
		}
		codes = append(codes, shared.VoucherCode{Code: code, PIN: pin})
	}
	return codes, nil
}

func (p *SelfProvider) Validate(ctx context.Context, code, pin string) (shared.ValidationResult, error) {
	return shared.ValidationResult{Valid: true}, nil
}

func (p *SelfProvider) Redeem(ctx context.Context, code, pin string, amount shared.Money) (shared.RedeemResult, error) {
	return shared.RedeemResult{Success: true, RedeemedAmount: amount}, nil
}

func randomCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	encoded = strings.ToUpper(encoded)
	if len(encoded) > n {
		encoded = encoded[:n]
	}
	return encoded, nil
}
