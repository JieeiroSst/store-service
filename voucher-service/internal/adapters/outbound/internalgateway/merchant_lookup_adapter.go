package internalgateway

import (
	"context"

	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type MerchantLookupAdapter struct {
	merchantSvc merchantapp.MerchantService
}

func NewMerchantLookupAdapter(merchantSvc merchantapp.MerchantService) voucherapp.MerchantLookup {
	return &MerchantLookupAdapter{merchantSvc: merchantSvc}
}

func (a *MerchantLookupAdapter) GetMerchantInfo(ctx context.Context, id shared.MerchantID) (voucherapp.MerchantInfo, error) {
	m, err := a.merchantSvc.GetMerchant(ctx, id)
	if err != nil {
		return voucherapp.MerchantInfo{}, err
	}
	return voucherapp.MerchantInfo{ID: m.ID, ProviderType: m.ProviderType, Active: m.IsActive()}, nil
}
