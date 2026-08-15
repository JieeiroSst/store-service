package app

import (
	auditapp "github.com/JIeeiroSst/voucher-service/internal/application/audit"
	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	corporateapp "github.com/JIeeiroSst/voucher-service/internal/application/corporate"
	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	fileapp "github.com/JIeeiroSst/voucher-service/internal/application/file"
	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
	reportingapp "github.com/JIeeiroSst/voucher-service/internal/application/reporting"
	schedulerapp "github.com/JIeeiroSst/voucher-service/internal/application/scheduler"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	walletapp "github.com/JIeeiroSst/voucher-service/internal/application/wallet"

	httpinbound "github.com/JIeeiroSst/voucher-service/internal/adapters/inbound/http"
	kafkaconsumer "github.com/JIeeiroSst/voucher-service/internal/adapters/inbound/consumer"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/authtoken"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/internalgateway"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/notifier"
	paymentgateway "github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/payment"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/postgres"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/provider"
	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/publisher"
	redisadapter "github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/redis"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"github.com/JIeeiroSst/voucher-service/internal/platform/consul"
	"github.com/JIeeiroSst/voucher-service/internal/platform/db"
	"github.com/JIeeiroSst/voucher-service/internal/platform/kafka"
	"github.com/JIeeiroSst/voucher-service/internal/platform/logger"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	platformredis "github.com/JIeeiroSst/voucher-service/internal/platform/redis"
	platformscheduler "github.com/JIeeiroSst/voucher-service/internal/platform/scheduler"
	"github.com/JIeeiroSst/voucher-service/internal/platform/tracing"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"

	"go.uber.org/fx"
)

// clockModule provides the one shared.Clock instance used across every
// application service — domain/shared has no fx of its own (pure domain
// layer), so this is the one place a system clock is wired in.
var clockModule = fx.Module("clock", fx.Provide(func() shared.Clock { return shared.NewSystemClock() }))

// Module composes every platform, application, and adapter sub-module so
// cmd/api/main.go only needs fx.New(app.Module).
var Module = fx.Options(
	// platform
	config.Module,
	logger.Module,
	db.Module,
	platformredis.Module,
	kafka.Module,
	txmanager.Module,
	tracing.Module,
	consul.Module,
	platformscheduler.Module,
	outbox.Module,
	clockModule,

	// application (driving/driven ports + orchestration)
	merchantapp.Module,
	voucherapp.Module,
	orderapp.Module,
	walletapp.Module,
	corporateapp.Module,
	distributionapp.Module,
	reconciliationapp.Module,
	paymentapp.Module,
	notificationapp.Module,
	inventoryapp.Module,
	auditapp.Module,
	reportingapp.Module,
	schedulerapp.Module,
	authapp.Module,
	fileapp.Module,
	partnerapp.Module,

	// outbound adapters
	postgres.Module,
	redisadapter.Module,
	provider.Module,
	publisher.Module,
	notifier.Module,
	paymentgateway.Module,
	authtoken.Module,
	internalgateway.Module,

	// inbound adapters
	httpinbound.Module,
	kafkaconsumer.Module,
)
