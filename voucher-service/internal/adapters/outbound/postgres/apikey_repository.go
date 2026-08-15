package postgres

import (
	"context"
	"encoding/json"
	"time"

	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	"gorm.io/gorm"
)

type apiKeyModel struct {
	ID              string     `gorm:"column:id;primaryKey"`
	PartnerID       string     `gorm:"column:partner_id"`
	KeyPrefix       string     `gorm:"column:key_prefix"`
	SecretHash      string     `gorm:"column:secret_hash"`
	Scopes          []byte     `gorm:"column:scopes"`
	RateLimitPerMin int        `gorm:"column:rate_limit_per_min"`
	Status          string     `gorm:"column:status"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (apiKeyModel) TableName() string { return "api_keys" }

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) partnerapp.APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *partnerapp.APIKey, encryptedSecret string) error {
	scopes, err := json.Marshal(key.Scopes)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	model := apiKeyModel{
		ID:              newUUID(),
		PartnerID:       key.PartnerID,
		KeyPrefix:       key.KeyPrefix,
		SecretHash:      encryptedSecret,
		Scopes:          scopes,
		RateLimitPerMin: key.RateLimitPerMin,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *APIKeyRepository) FindByPrefix(ctx context.Context, keyPrefix string) (*partnerapp.APIKey, error) {
	var m apiKeyModel
	if err := r.db.WithContext(ctx).First(&m, "key_prefix = ?", keyPrefix).Error; err != nil {
		return nil, err
	}
	var scopes []string
	_ = json.Unmarshal(m.Scopes, &scopes)
	return &partnerapp.APIKey{
		ID:              m.ID,
		PartnerID:       m.PartnerID,
		KeyPrefix:       m.KeyPrefix,
		EncryptedSecret: m.SecretHash,
		Scopes:          scopes,
		RateLimitPerMin: m.RateLimitPerMin,
		Status:          m.Status,
		LastUsedAt:      m.LastUsedAt,
	}, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, keyPrefix string) error {
	return r.db.WithContext(ctx).Model(&apiKeyModel{}).
		Where("key_prefix = ?", keyPrefix).
		Updates(map[string]any{"status": "revoked", "updated_at": time.Now().UTC()}).Error
}

func (r *APIKeyRepository) TouchLastUsed(ctx context.Context, keyPrefix string) error {
	return r.db.WithContext(ctx).Model(&apiKeyModel{}).
		Where("key_prefix = ?", keyPrefix).
		Update("last_used_at", time.Now().UTC()).Error
}
