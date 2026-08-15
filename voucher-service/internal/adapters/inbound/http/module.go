package http

import (
	"fmt"
	stdhttp "net/http"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	redisadapter "github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/redis"
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/server"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newHandlers(
	voucher *VoucherHandler, order *OrderHandler, merchant *MerchantHandler, wallet *WalletHandler,
	corporate *CorporateHandler, auth *AuthHandler, distribution *DistributionHandler,
	payment *PaymentHandler, partner *PartnerHandler, file *FileHandler, reporting *ReportingHandler,
) Handlers {
	return Handlers{
		Voucher: voucher, Order: order, Merchant: merchant, Wallet: wallet, Corporate: corporate,
		Auth: auth, Distribution: distribution, Payment: payment, Partner: partner, File: file, Reporting: reporting,
	}
}

func newEngine(
	handlers Handlers, authSvc authapp.AuthService, partnerSvc partnerapp.PartnerService,
	rateLimiter *redisadapter.RateLimiter, idempStore idempotency.Store, tracer trace.Tracer, log *zap.Logger,
) *gin.Engine {
	return NewRouter(RouterDeps{
		Handlers: handlers, AuthSvc: authSvc, PartnerSvc: partnerSvc,
		RateLimiter: rateLimiter, IdempStore: idempStore, Tracer: tracer, Log: log,
	})
}

func newHTTPServer(cfg *config.Config, engine *gin.Engine) *stdhttp.Server {
	return &stdhttp.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: engine,
	}
}

var Module = fx.Module("http-inbound",
	fx.Provide(
		NewVoucherHandler, NewOrderHandler, NewMerchantHandler, NewWalletHandler,
		NewCorporateHandler, NewAuthHandler, NewDistributionHandler, NewPaymentHandler,
		NewPartnerHandler, NewFileHandler, NewReportingHandler,
		newHandlers, newEngine, newHTTPServer,
	),
	fx.Invoke(server.Serve),
)
