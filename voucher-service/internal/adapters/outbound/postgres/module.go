package postgres

import (
	auditapp "github.com/JIeeiroSst/voucher-service/internal/application/audit"
	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	corporateapp "github.com/JIeeiroSst/voucher-service/internal/application/corporate"
	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
	reportingapp "github.com/JIeeiroSst/voucher-service/internal/application/reporting"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	walletapp "github.com/JIeeiroSst/voucher-service/internal/application/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("postgres-adapters",
	fx.Provide(
		fx.Annotate(NewMerchantRepository, fx.As(new(merchantapp.MerchantRepository))),
		fx.Annotate(NewVoucherRepository, fx.As(new(voucherapp.VoucherRepository))),
		fx.Annotate(NewVoucherStockRepository, fx.As(new(inventoryapp.StockRepository))),
		fx.Annotate(NewVoucherStockRepository, fx.As(new(inventoryapp.StockClaimer))),
		fx.Annotate(NewOrderRepository, fx.As(new(orderapp.OrderRepository))),
		fx.Annotate(NewWalletRepository, fx.As(new(walletapp.WalletRepository))),
		fx.Annotate(NewLedgerRepository, fx.As(new(walletapp.LedgerRepository))),
		fx.Annotate(NewCorporateRepository, fx.As(new(corporateapp.CorporateRepository))),
		fx.Annotate(NewIdempotencyRepository, fx.As(new(idempotency.Store))),
		fx.Annotate(NewOutboxRepository, fx.As(new(outbox.Outbox))),
		fx.Annotate(NewOutboxRepository, fx.As(new(outbox.Repository))),
		fx.Annotate(NewAuditRepository, fx.As(new(auditapp.AuditRepository))),
		fx.Annotate(NewNotificationRepository, fx.As(new(notificationapp.NotificationRepository))),
		fx.Annotate(NewPaymentRepository, fx.As(new(paymentapp.PaymentRepository))),
		fx.Annotate(NewAPIKeyRepository, fx.As(new(partnerapp.APIKeyRepository))),
		fx.Annotate(NewJobRepository, fx.As(new(distributionapp.JobRepository))),
		fx.Annotate(NewClaimRepository, fx.As(new(distributionapp.ClaimRepository))),
		fx.Annotate(NewRunRepository, fx.As(new(reconciliationapp.RunRepository))),
		fx.Annotate(NewReportingRepository, fx.As(new(reportingapp.ReportingRepository))),
		fx.Annotate(NewUserRepository, fx.As(new(authapp.UserRepository))),
	),
)
