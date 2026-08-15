package internalgateway

import (
	"context"

	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type VoucherBulkIssuerAdapter struct {
	voucherSvc voucherapp.VoucherService
}

func NewVoucherBulkIssuerAdapter(voucherSvc voucherapp.VoucherService) distributionapp.VoucherBulkIssuer {
	return &VoucherBulkIssuerAdapter{voucherSvc: voucherSvc}
}

func (a *VoucherBulkIssuerAdapter) IssueBulk(ctx context.Context, merchantID shared.MerchantID, productSKU string, denomination shared.Money, quantity int, corporateID shared.CorporateID) ([]shared.VoucherID, error) {
	issued, err := a.voucherSvc.IssueVouchers(ctx, voucherapp.IssueVouchersInput{
		MerchantID:   merchantID,
		ProductSKU:   productSKU,
		Denomination: denomination,
		Quantity:     quantity,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]shared.VoucherID, len(issued))
	for i, iv := range issued {
		ids[i] = iv.Voucher.ID
	}
	return ids, nil
}
