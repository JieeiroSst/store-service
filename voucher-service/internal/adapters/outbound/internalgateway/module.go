package internalgateway

import (
	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	fileapp "github.com/JIeeiroSst/voucher-service/internal/application/file"
	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
	schedulerapp "github.com/JIeeiroSst/voucher-service/internal/application/scheduler"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"go.uber.org/fx"
)

var Module = fx.Module("internalgateway",
	fx.Provide(
		fx.Annotate(NewVoucherIssuerAdapter, fx.As(new(orderapp.VoucherIssuer))),
		fx.Annotate(NewPaymentInitiatorAdapter, fx.As(new(orderapp.PaymentInitiator))),
		fx.Annotate(NewWalletDebiterAdapter, fx.As(new(orderapp.WalletDebiter))),
		fx.Annotate(NewMerchantLookupAdapter, fx.As(new(voucherapp.MerchantLookup))),
		fx.Annotate(NewVoucherBulkIssuerAdapter, fx.As(new(distributionapp.VoucherBulkIssuer))),
		fx.Annotate(NewBudgetCheckerAdapter, fx.As(new(distributionapp.BudgetChecker))),
		fx.Annotate(NewVoucherExpirerAdapter, fx.As(new(schedulerapp.VoucherExpirer))),
		fx.Annotate(NewReconciliationTriggerAdapter, fx.As(new(schedulerapp.ReconciliationTrigger))),
		fx.Annotate(NewLowStockAlerterAdapter, fx.As(new(schedulerapp.LowStockAlerter))),
		fx.Annotate(NewPaymentRecordSourceAdapter, fx.As(new(reconciliationapp.PaymentRecordSource))),
		fx.Annotate(NewStockImporterAdapter, fx.As(new(fileapp.StockImporter))),
		fx.Annotate(NewReportSourceAdapter, fx.As(new(fileapp.ReportSource))),
	),
)
