package merchant

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Merchant struct {
	ID               shared.MerchantID
	Name             string
	ProviderType     shared.ProviderType
	Config           map[string]any
	Status           Status
	Version          int
	PersistedVersion int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewMerchant(name string, providerType shared.ProviderType, config map[string]any, now time.Time) (*Merchant, error) {
	if name == "" {
		return nil, ErrInvalidMerchant
	}
	if !providerType.Valid() {
		return nil, ErrUnsupportedProviderType
	}
	if config == nil {
		config = map[string]any{}
	}
	return &Merchant{
		ID:           shared.NewMerchantID(),
		Name:         name,
		ProviderType: providerType,
		Config:       config,
		Status:       StatusActive,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (m *Merchant) Activate(now time.Time) {
	m.Status = StatusActive
	m.Version++
	m.UpdatedAt = now
}

func (m *Merchant) Deactivate(now time.Time) {
	m.Status = StatusInactive
	m.Version++
	m.UpdatedAt = now
}

func (m *Merchant) IsActive() bool {
	return m.Status == StatusActive
}

func (m *Merchant) UpdateConfig(config map[string]any, now time.Time) {
	m.Config = config
	m.Version++
	m.UpdatedAt = now
}
