package corporate

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Corporate struct {
	ID               shared.CorporateID
	Name             string
	TaxCode          string
	ContactEmail     string
	BudgetLimit      *shared.Money
	Status           Status
	Version          int
	PersistedVersion int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewCorporate(name, taxCode, contactEmail string, budgetLimit *shared.Money, now time.Time) (*Corporate, error) {
	if name == "" {
		return nil, ErrInvalidCorporate
	}
	return &Corporate{
		ID:           shared.NewCorporateID(),
		Name:         name,
		TaxCode:      taxCode,
		ContactEmail: contactEmail,
		BudgetLimit:  budgetLimit,
		Status:       StatusActive,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (c *Corporate) CheckBudget(alreadySpent, proposedSpend shared.Money) error {
	if c.BudgetLimit == nil {
		return nil
	}
	total, err := alreadySpent.Add(proposedSpend)
	if err != nil {
		return err
	}
	if c.BudgetLimit.LessThan(total) {
		return ErrBudgetExceeded
	}
	return nil
}

func (c *Corporate) SetBudget(limit shared.Money, now time.Time) {
	c.BudgetLimit = &limit
	c.Version++
	c.UpdatedAt = now
}

func (c *Corporate) IsActive() bool { return c.Status == StatusActive }
