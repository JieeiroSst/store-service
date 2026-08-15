package postgres

import (
	"context"
	"time"

	corporateapp "github.com/JIeeiroSst/voucher-service/internal/application/corporate"
	"github.com/JIeeiroSst/voucher-service/internal/domain/corporate"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type corporateModel struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Name           string    `gorm:"column:name"`
	TaxCode        string    `gorm:"column:tax_code"`
	ContactEmail   string    `gorm:"column:contact_email"`
	BudgetLimit    *float64  `gorm:"column:budget_limit"`
	BudgetCurrency *string   `gorm:"column:budget_currency"`
	Status         string    `gorm:"column:status"`
	Version        int       `gorm:"column:version"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (corporateModel) TableName() string { return "corporates" }

func corporateToModel(c *corporate.Corporate) *corporateModel {
	m := &corporateModel{
		ID:           c.ID.String(),
		Name:         c.Name,
		TaxCode:      c.TaxCode,
		ContactEmail: c.ContactEmail,
		Status:       string(c.Status),
		Version:      c.Version,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if c.BudgetLimit != nil {
		amt := float64(c.BudgetLimit.Amount)
		m.BudgetLimit = &amt
		m.BudgetCurrency = &c.BudgetLimit.Currency
	}
	return m
}

func corporateFromModel(m *corporateModel) (*corporate.Corporate, error) {
	id, err := shared.ParseCorporateID(m.ID)
	if err != nil {
		return nil, err
	}
	c := &corporate.Corporate{
		ID:               id,
		Name:             m.Name,
		TaxCode:          m.TaxCode,
		ContactEmail:     m.ContactEmail,
		Status:           corporate.Status(m.Status),
		Version:          m.Version,
		PersistedVersion: m.Version,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.BudgetLimit != nil && m.BudgetCurrency != nil {
		limit := shared.NewMoney(int64(*m.BudgetLimit), *m.BudgetCurrency)
		c.BudgetLimit = &limit
	}
	return c, nil
}

type CorporateRepository struct {
	db *gorm.DB
}

func NewCorporateRepository(db *gorm.DB) corporateapp.CorporateRepository {
	return &CorporateRepository{db: db}
}

func (r *CorporateRepository) Create(ctx context.Context, c *corporate.Corporate) error {
	return txmanager.DBFromContext(ctx, r.db).Create(corporateToModel(c)).Error
}

func (r *CorporateRepository) FindByID(ctx context.Context, id shared.CorporateID) (*corporate.Corporate, error) {
	var m corporateModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&m, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, corporate.ErrCorporateNotFound
		}
		return nil, err
	}
	return corporateFromModel(&m)
}

func (r *CorporateRepository) Save(ctx context.Context, c *corporate.Corporate) error {
	model := corporateToModel(c)
	tx := txmanager.DBFromContext(ctx, r.db).Model(&corporateModel{}).
		Where("id = ? AND version = ?", model.ID, c.PersistedVersion).
		Updates(map[string]any{
			"budget_limit":    model.BudgetLimit,
			"budget_currency": model.BudgetCurrency,
			"status":          model.Status,
			"version":         model.Version,
			"updated_at":      model.UpdatedAt,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return corporate.ErrVersionConflict
	}
	return nil
}

func (r *CorporateRepository) SpentThisPeriod(ctx context.Context, id shared.CorporateID) (shared.Money, error) {
	var total float64
	err := txmanager.DBFromContext(ctx, r.db).
		Table("wallet_transactions wt").
		Joins("JOIN wallets w ON w.id = wt.wallet_id").
		Where("w.owner_type = 'corporate' AND w.owner_id = ? AND wt.type = 'debit'", id.String()).
		Select("COALESCE(SUM(wt.amount), 0)").
		Scan(&total).Error
	if err != nil {
		return shared.Money{}, err
	}
	return shared.NewMoney(int64(total), "VND"), nil
}
