package internalgateway

import (
	"context"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
)

type VoucherIssuerAdapter struct {
	voucherSvc voucherapp.VoucherService
}

func NewVoucherIssuerAdapter(voucherSvc voucherapp.VoucherService) orderapp.VoucherIssuer {
	return &VoucherIssuerAdapter{voucherSvc: voucherSvc}
}

func (a *VoucherIssuerAdapter) IssueVouchersForOrder(ctx context.Context, req orderapp.VoucherIssuanceRequest) ([]orderapp.IssuedVoucherRef, error) {
	var refs []orderapp.IssuedVoucherRef
	for _, item := range req.Items {
		orderID := req.OrderID
		issued, err := a.voucherSvc.IssueVouchers(ctx, voucherapp.IssueVouchersInput{
			MerchantID:   item.MerchantID,
			ProductSKU:   item.ProductSKU,
			Denomination: item.Denomination,
			Quantity:     item.Quantity,
			OrderID:      &orderID,
		})
		if err != nil {
			return nil, err
		}
		for _, iv := range issued {
			refs = append(refs, orderapp.IssuedVoucherRef{
				VoucherID:  iv.Voucher.ID.String(),
				Code:       iv.Voucher.Code,
				MerchantID: iv.Voucher.MerchantID.String(),
				PIN:        iv.PlaintextPIN,
			})
		}
	}
	return refs, nil
}
