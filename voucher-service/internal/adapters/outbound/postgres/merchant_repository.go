package postgres

import (
	"context"
	"encoding/json"
	"time"

	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type merchantModel struct {
	ID           string `gorm:"column:id;primaryKey"`
	Name         string `gorm:"column:name"`
	ProviderType string `gorm:"column:provider_type"`
	Config       []byte `gorm:"column:config"`
	Status       string `gorm:"column:status"`
	Version      int    `gorm:"column:version"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (merchantModel) TableName() string { return "merchants" }

func merchantToModel(m *merchant.Merchant) (*merchantModel, error) {
	cfg, err := json.Marshal(m.Config)
	if err != nil {
		return nil, err
	}
	return &merchantModel{
		ID:           m.ID.String(),
		Name:         m.Name,
		ProviderType: string(m.ProviderType),
		Config:       cfg,
		Status:       string(m.Status),
		Version:      m.Version,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func merchantFromModel(m *merchantModel) (*merchant.Merchant, error) {
	id, err := shared.ParseMerchantID(m.ID)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if len(m.Config) > 0 {
		if err := json.Unmarshal(m.Config, &cfg); err != nil {
			return nil, err
		}
	}
	return &merchant.Merchant{
		ID:               id,
		Name:             m.Name,
		ProviderType:     shared.ProviderType(m.ProviderType),
		Config:           cfg,
		Status:           merchant.Status(m.Status),
		Version:          m.Version,
		PersistedVersion: m.Version,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

type MerchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) merchantapp.MerchantRepository {
	return &MerchantRepository{db: db}
}

func (r *MerchantRepository) Create(ctx context.Context, m *merchant.Merchant) error {
	model, err := merchantToModel(m)
	if err != nil {
		return err
	}
	return txmanager.DBFromContext(ctx, r.db).Create(model).Error
}

func (r *MerchantRepository) FindByID(ctx context.Context, id shared.MerchantID) (*merchant.Merchant, error) {
	var model merchantModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&model, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, merchant.ErrMerchantNotFound
		}
		return nil, err
	}
	return merchantFromModel(&model)
}

func (r *MerchantRepository) FindAll(ctx context.Context) ([]*merchant.Merchant, error) {
	var models []merchantModel
	if err := txmanager.DBFromContext(ctx, r.db).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*merchant.Merchant, 0, len(models))
	for i := range models {
		m, err := merchantFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *MerchantRepository) Save(ctx context.Context, m *merchant.Merchant) error {
	model, err := merchantToModel(m)
	if err != nil {
		return err
	}
	tx := txmanager.DBFromContext(ctx, r.db).Model(&merchantModel{}).
		Where("id = ? AND version = ?", model.ID, m.PersistedVersion).
		Updates(map[string]any{
			"name":          model.Name,
			"provider_type": model.ProviderType,
			"config":        model.Config,
			"status":        model.Status,
			"version":       model.Version,
			"updated_at":    model.UpdatedAt,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return merchant.ErrVersionConflict
	}
	return nil
}
