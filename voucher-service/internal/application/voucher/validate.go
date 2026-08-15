package voucher

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

func (s *Service) ValidateVoucher(ctx context.Context, in ValidateVoucherInput) (shared.ValidationResult, error) {
	v, err := s.repo.FindByID(ctx, in.VoucherID)
	if err != nil {
		return shared.ValidationResult{}, err
	}

	if err := v.ValidatePIN(in.PIN); err != nil {
		return shared.ValidationResult{Valid: false, Reason: err.Error()}, nil
	}
	if !v.CanRedeem(s.clock.Now()) {
		return shared.ValidationResult{Valid: false, Reason: "voucher not redeemable in status " + string(v.Status)}, nil
	}

	info, err := s.merchantLookup.GetMerchantInfo(ctx, v.MerchantID)
	if err != nil {
		return shared.ValidationResult{}, err
	}
	if info.ProviderType == shared.ProviderTypeAPI {
		provider, err := s.registry.Resolve(shared.ProviderTypeAPI)
		if err != nil {
			return shared.ValidationResult{}, err
		}
		return provider.Validate(ctx, v.Code, in.PIN)
	}

	return shared.ValidationResult{Valid: true, Balance: v.ProductRef.Denomination}, nil
}
