package provider

import (
	"fmt"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Registry struct {
	byType map[shared.ProviderType]voucherapp.MerchantProvider
}

func NewRegistry(providers []voucherapp.MerchantProvider) voucherapp.ProviderRegistry {
	byType := make(map[shared.ProviderType]voucherapp.MerchantProvider, len(providers))
	for _, p := range providers {
		byType[p.Type()] = p
	}
	return &Registry{byType: byType}
}

func (r *Registry) Resolve(providerType shared.ProviderType) (voucherapp.MerchantProvider, error) {
	p, ok := r.byType[providerType]
	if !ok {
		return nil, fmt.Errorf("no provider registered for type %q", providerType)
	}
	return p, nil
}
