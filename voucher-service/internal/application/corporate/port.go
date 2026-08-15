package corporate

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/corporate"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type RegisterCorporateInput struct {
	Name         string
	TaxCode      string
	ContactEmail string
	BudgetLimit  *shared.Money
}

type CorporateService interface {
	RegisterCorporate(ctx context.Context, in RegisterCorporateInput) (*corporate.Corporate, error)
	GetCorporate(ctx context.Context, id shared.CorporateID) (*corporate.Corporate, error)
	SetBudget(ctx context.Context, id shared.CorporateID, limit shared.Money) (*corporate.Corporate, error)
	CheckBudget(ctx context.Context, id shared.CorporateID, proposedSpend shared.Money) error
}

type CorporateRepository interface {
	Create(ctx context.Context, c *corporate.Corporate) error
	FindByID(ctx context.Context, id shared.CorporateID) (*corporate.Corporate, error)
	Save(ctx context.Context, c *corporate.Corporate) error
	SpentThisPeriod(ctx context.Context, id shared.CorporateID) (shared.Money, error)
}
