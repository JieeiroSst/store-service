package postgres

import (
	"context"
	"time"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type paymentModel struct {
	ID             string  `gorm:"column:id;primaryKey"`
	OrderID        string  `gorm:"column:order_id"`
	Provider       string  `gorm:"column:provider"`
	Amount         float64 `gorm:"column:amount"`
	Currency       string  `gorm:"column:currency"`
	Status         string  `gorm:"column:status"`
	ProviderTxnRef *string `gorm:"column:provider_txn_ref"`
	IdempotencyKey *string `gorm:"column:idempotency_key"`
	Signature      *string `gorm:"column:signature"`
	RawCallback    []byte  `gorm:"column:raw_callback"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (paymentModel) TableName() string { return "payments" }

func paymentToModel(p *paymentapp.Payment) *paymentModel {
	m := &paymentModel{
		ID:        p.ID,
		OrderID:   p.OrderID.String(),
		Provider:  p.Provider,
		Amount:    float64(p.Amount.Amount),
		Currency:  p.Amount.Currency,
		Status:    string(p.Status),
		UpdatedAt: time.Now().UTC(),
	}
	if p.ProviderTxnRef != "" {
		m.ProviderTxnRef = &p.ProviderTxnRef
	}
	return m
}

func paymentFromModel(m *paymentModel) (*paymentapp.Payment, error) {
	orderID, err := shared.ParseOrderID(m.OrderID)
	if err != nil {
		return nil, err
	}
	p := &paymentapp.Payment{
		ID:       m.ID,
		OrderID:  orderID,
		Provider: m.Provider,
		Amount:   shared.NewMoney(int64(m.Amount), m.Currency),
		Status:   paymentapp.Status(m.Status),
	}
	if m.ProviderTxnRef != nil {
		p.ProviderTxnRef = *m.ProviderTxnRef
	}
	return p, nil
}

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) paymentapp.PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *paymentapp.Payment) error {
	if p.ID == "" {
		p.ID = newUUID()
	}
	model := paymentToModel(p)
	model.CreatedAt = time.Now().UTC()
	return txmanager.DBFromContext(ctx, r.db).Create(model).Error
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*paymentapp.Payment, error) {
	var m paymentModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return paymentFromModel(&m)
}

func (r *PaymentRepository) FindByProviderTxnRef(ctx context.Context, ref string) (*paymentapp.Payment, error) {
	var m paymentModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&m, "provider_txn_ref = ?", ref).Error; err != nil {
		return nil, err
	}
	return paymentFromModel(&m)
}

func (r *PaymentRepository) ListSettledSince(ctx context.Context, since time.Time) ([]*paymentapp.Payment, error) {
	var models []paymentModel
	err := txmanager.DBFromContext(ctx, r.db).
		Where("status = 'succeeded' AND updated_at >= ?", since).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*paymentapp.Payment, 0, len(models))
	for i := range models {
		p, err := paymentFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PaymentRepository) Save(ctx context.Context, p *paymentapp.Payment) error {
	model := paymentToModel(p)
	return txmanager.DBFromContext(ctx, r.db).Model(&paymentModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"status":           model.Status,
			"provider_txn_ref": model.ProviderTxnRef,
			"updated_at":       model.UpdatedAt,
		}).Error
}
