package merchant

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type RegisterMerchantInput struct {
	Name         string
	ProviderType shared.ProviderType
	Config       map[string]any
}

type MerchantService interface {
	RegisterMerchant(ctx context.Context, in RegisterMerchantInput) (*merchant.Merchant, error)
	GetMerchant(ctx context.Context, id shared.MerchantID) (*merchant.Merchant, error)
	ListMerchants(ctx context.Context) ([]*merchant.Merchant, error)
	UpdateMerchantConfig(ctx context.Context, id shared.MerchantID, config map[string]any) (*merchant.Merchant, error)
	DeactivateMerchant(ctx context.Context, id shared.MerchantID) error
}

type MerchantRepository interface {
	Create(ctx context.Context, m *merchant.Merchant) error
	FindByID(ctx context.Context, id shared.MerchantID) (*merchant.Merchant, error)
	FindAll(ctx context.Context) ([]*merchant.Merchant, error)
	Save(ctx context.Context, m *merchant.Merchant) error
}
