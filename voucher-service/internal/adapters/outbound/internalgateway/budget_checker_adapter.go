package internalgateway

import (
	"context"

	corporateapp "github.com/JIeeiroSst/voucher-service/internal/application/corporate"
	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type BudgetCheckerAdapter struct {
	corporateSvc corporateapp.CorporateService
}

func NewBudgetCheckerAdapter(corporateSvc corporateapp.CorporateService) distributionapp.BudgetChecker {
	return &BudgetCheckerAdapter{corporateSvc: corporateSvc}
}

func (a *BudgetCheckerAdapter) CheckBudget(ctx context.Context, corporateID shared.CorporateID, proposedSpend shared.Money) error {
	return a.corporateSvc.CheckBudget(ctx, corporateID, proposedSpend)
}
