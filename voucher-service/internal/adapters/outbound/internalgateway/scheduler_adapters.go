package internalgateway

import (
	"context"

	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
	schedulerapp "github.com/JIeeiroSst/voucher-service/internal/application/scheduler"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
)

type VoucherExpirerAdapter struct {
	voucherSvc voucherapp.VoucherService
}

func NewVoucherExpirerAdapter(voucherSvc voucherapp.VoucherService) schedulerapp.VoucherExpirer {
	return &VoucherExpirerAdapter{voucherSvc: voucherSvc}
}

func (a *VoucherExpirerAdapter) ExpireDueVouchers(ctx context.Context) (int, error) {
	return a.voucherSvc.ExpireDueVouchers(ctx)
}

type ReconciliationTriggerAdapter struct {
	reconciliationSvc reconciliationapp.ReconciliationService
}

func NewReconciliationTriggerAdapter(reconciliationSvc reconciliationapp.ReconciliationService) schedulerapp.ReconciliationTrigger {
	return &ReconciliationTriggerAdapter{reconciliationSvc: reconciliationSvc}
}

func (a *ReconciliationTriggerAdapter) RunPaymentReconciliation(ctx context.Context) error {
	_, err := a.reconciliationSvc.RunPaymentReconciliation(ctx)
	return err
}

type LowStockAlerterAdapter struct {
	inventorySvc     inventoryapp.InventoryService
	notificationSvc  notificationapp.NotificationService
}

func NewLowStockAlerterAdapter(inventorySvc inventoryapp.InventoryService, notificationSvc notificationapp.NotificationService) schedulerapp.LowStockAlerter {
	return &LowStockAlerterAdapter{inventorySvc: inventorySvc, notificationSvc: notificationSvc}
}

func (a *LowStockAlerterAdapter) CheckAndAlertLowStock(ctx context.Context, threshold int) error {
	levels, err := a.inventorySvc.ListLowStock(ctx, threshold)
	if err != nil {
		return err
	}
	for _, lvl := range levels {
		err := a.notificationSvc.Send(ctx, notificationapp.SendInput{
			RecipientType: "system",
			RecipientID:   "ops-team",
			Channel:       notificationapp.ChannelEmail,
			TemplateCode:  "low_stock_alert",
			Payload: map[string]any{
				"merchant_id":     lvl.MerchantID.String(),
				"product_sku":     lvl.ProductSKU,
				"available_count": lvl.AvailableCount,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
