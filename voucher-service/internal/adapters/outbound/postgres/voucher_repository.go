package postgres

import (
	"context"
	"time"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type voucherModel struct {
	ID              string     `gorm:"column:id;primaryKey"`
	MerchantID      string     `gorm:"column:merchant_id"`
	OrderID         *string    `gorm:"column:order_id"`
	OwnerType       *string    `gorm:"column:owner_type"`
	OwnerID         *string    `gorm:"column:owner_id"`
	ProductSKU      string     `gorm:"column:product_sku"`
	Denomination    float64    `gorm:"column:denomination"`
	Currency        string     `gorm:"column:currency"`
	Code            string     `gorm:"column:code"`
	PinHash         string     `gorm:"column:pin_hash"`
	Status          string     `gorm:"column:status"`
	Version         int        `gorm:"column:version"`
	IdempotencyKey  *string    `gorm:"column:idempotency_key"`
	RedeemedAmount  *float64   `gorm:"column:redeemed_amount"`
	ProviderTxnRef  string     `gorm:"column:provider_txn_ref"`
	IssuedAt        *time.Time `gorm:"column:issued_at"`
	ActivatedAt     *time.Time `gorm:"column:activated_at"`
	RedeemedAt      *time.Time `gorm:"column:redeemed_at"`
	ExpiresAt       *time.Time `gorm:"column:expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (voucherModel) TableName() string { return "vouchers" }

func voucherToModel(v *voucher.Voucher) *voucherModel {
	m := &voucherModel{
		ID:             v.ID.String(),
		MerchantID:     v.MerchantID.String(),
		ProductSKU:     v.ProductRef.SKU,
		Denomination:   float64(v.ProductRef.Denomination.Amount),
		Currency:       v.ProductRef.Denomination.Currency,
		Code:           v.Code,
		PinHash:        v.PinHash,
		Status:         string(v.Status),
		Version:        v.Version,
		ProviderTxnRef: v.ProviderTxnRef,
		IssuedAt:       v.IssuedAt,
		ActivatedAt:    v.ActivatedAt,
		RedeemedAt:     v.RedeemedAt,
		ExpiresAt:      v.ExpiresAt,
		RevokedAt:      v.RevokedAt,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
	if v.OrderID != nil {
		s := v.OrderID.String()
		m.OrderID = &s
	}
	if v.OwnerID != nil {
		ot := string(v.OwnerType)
		m.OwnerType = &ot
		m.OwnerID = v.OwnerID
	}
	if v.IdempotencyKey != "" {
		m.IdempotencyKey = &v.IdempotencyKey
	}
	if v.RedeemedAmount != nil {
		amt := float64(v.RedeemedAmount.Amount)
		m.RedeemedAmount = &amt
	}
	return m
}

func voucherFromModel(m *voucherModel) (*voucher.Voucher, error) {
	id, err := shared.ParseVoucherID(m.ID)
	if err != nil {
		return nil, err
	}
	merchantID, err := shared.ParseMerchantID(m.MerchantID)
	if err != nil {
		return nil, err
	}
	v := &voucher.Voucher{
		ID:         id,
		MerchantID: merchantID,
		ProductRef: shared.ProductRef{
			MerchantID:   merchantID,
			SKU:          m.ProductSKU,
			Denomination: shared.NewMoney(int64(m.Denomination), m.Currency),
		},
		Code:             m.Code,
		PinHash:          m.PinHash,
		Status:           voucher.Status(m.Status),
		Version:          m.Version,
		PersistedVersion: m.Version,
		ProviderTxnRef:   m.ProviderTxnRef,
		IssuedAt:       m.IssuedAt,
		ActivatedAt:    m.ActivatedAt,
		RedeemedAt:     m.RedeemedAt,
		ExpiresAt:      m.ExpiresAt,
		RevokedAt:      m.RevokedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.OrderID != nil {
		oid, err := shared.ParseOrderID(*m.OrderID)
		if err != nil {
			return nil, err
		}
		v.OrderID = &oid
	}
	if m.OwnerType != nil {
		v.OwnerType = voucher.OwnerType(*m.OwnerType)
		v.OwnerID = m.OwnerID
	}
	if m.IdempotencyKey != nil {
		v.IdempotencyKey = *m.IdempotencyKey
	}
	if m.RedeemedAmount != nil {
		amt := shared.NewMoney(int64(*m.RedeemedAmount), m.Currency)
		v.RedeemedAmount = &amt
	}
	return v, nil
}

type VoucherRepository struct {
	db *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) voucherapp.VoucherRepository {
	return &VoucherRepository{db: db}
}

func (r *VoucherRepository) Create(ctx context.Context, v *voucher.Voucher) error {
	return txmanager.DBFromContext(ctx, r.db).Create(voucherToModel(v)).Error
}

func (r *VoucherRepository) FindByID(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error) {
	var m voucherModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&m, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, voucher.ErrVoucherNotFound
		}
		return nil, err
	}
	return voucherFromModel(&m)
}

// FindByIDForUpdate takes SELECT ... FOR UPDATE, and must be called from
// within an active transaction (txmanager.WithinTx) so the lock holds for
// the duration of the caller's critical section.
func (r *VoucherRepository) FindByIDForUpdate(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error) {
	var m voucherModel
	db := txmanager.DBFromContext(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.First(&m, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, voucher.ErrVoucherNotFound
		}
		return nil, err
	}
	return voucherFromModel(&m)
}

func (r *VoucherRepository) FindByCode(ctx context.Context, merchantID shared.MerchantID, code string) (*voucher.Voucher, error) {
	var m voucherModel
	err := txmanager.DBFromContext(ctx, r.db).
		First(&m, "merchant_id = ? AND code = ?", merchantID.String(), code).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, voucher.ErrVoucherNotFound
		}
		return nil, err
	}
	return voucherFromModel(&m)
}

func (r *VoucherRepository) ListByOwner(ctx context.Context, ownerType voucher.OwnerType, ownerID string) ([]*voucher.Voucher, error) {
	var models []voucherModel
	err := txmanager.DBFromContext(ctx, r.db).
		Where("owner_type = ? AND owner_id = ?", string(ownerType), ownerID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*voucher.Voucher, 0, len(models))
	for i := range models {
		v, err := voucherFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *VoucherRepository) ListDueForExpiry(ctx context.Context, now time.Time) ([]*voucher.Voucher, error) {
	var models []voucherModel
	err := txmanager.DBFromContext(ctx, r.db).
		Where("status IN ('issued','active') AND expires_at IS NOT NULL AND expires_at <= ?", now).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*voucher.Voucher, 0, len(models))
	for i := range models {
		v, err := voucherFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// Save persists with an optimistic-lock guard: the WHERE clause targets
// v.PersistedVersion (the version as loaded), not v.Version - 1 — a
// mutation sequence may have bumped Version more than once in memory
// since the aggregate was loaded (e.g. Redeem plus a subsequent Revoke in
// the same call), so only the as-loaded version is a safe guard.
func (r *VoucherRepository) Save(ctx context.Context, v *voucher.Voucher) error {
	model := voucherToModel(v)
	tx := txmanager.DBFromContext(ctx, r.db).Model(&voucherModel{}).
		Where("id = ? AND version = ?", model.ID, v.PersistedVersion).
		Updates(map[string]any{
			"order_id":         model.OrderID,
			"owner_type":       model.OwnerType,
			"owner_id":         model.OwnerID,
			"code":             model.Code,
			"pin_hash":         model.PinHash,
			"status":           model.Status,
			"version":          model.Version,
			"idempotency_key":  model.IdempotencyKey,
			"redeemed_amount":  model.RedeemedAmount,
			"provider_txn_ref": model.ProviderTxnRef,
			"issued_at":        model.IssuedAt,
			"activated_at":     model.ActivatedAt,
			"redeemed_at":      model.RedeemedAt,
			"expires_at":       model.ExpiresAt,
			"revoked_at":       model.RevokedAt,
			"updated_at":       model.UpdatedAt,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return voucher.ErrVersionConflict
	}
	return nil
}
