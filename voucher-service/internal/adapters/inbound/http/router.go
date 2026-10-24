package http

import (
	"github.com/JIeeiroSst/voucher-service/internal/adapters/inbound/http/middleware"
	redisadapter "github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/redis"
	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Handlers struct {
	Voucher      *VoucherHandler
	Order        *OrderHandler
	Merchant     *MerchantHandler
	Wallet       *WalletHandler
	Corporate    *CorporateHandler
	Distribution *DistributionHandler
	Payment      *PaymentHandler
	Partner      *PartnerHandler
	File         *FileHandler
	Reporting    *ReportingHandler
}

type RouterDeps struct {
	Handlers    Handlers
	AuthSvc     authapp.AuthService
	PartnerSvc  partnerapp.PartnerService
	RateLimiter *redisadapter.RateLimiter
	IdempStore  idempotency.Store
	Tracer      trace.Tracer
	Log         *zap.Logger
}

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Tracing(deps.Tracer), middleware.StructuredLogger(deps.Log), gin.Recovery())

	r.GET("/health", Health)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/payments/:provider/webhook", deps.Handlers.Payment.Webhook)

		authorized := v1.Group("")
		authorized.Use(middleware.Auth(deps.AuthSvc))
		{
			mutate := authorized.Group("")
			mutate.Use(middleware.Idempotency(deps.IdempStore))
			{
				mutate.POST("/vouchers/:id/redeem", deps.Handlers.Voucher.Redeem)
				mutate.POST("/orders", deps.Handlers.Order.Create)
				mutate.POST("/orders/:id/checkout", deps.Handlers.Order.Checkout)
				mutate.POST("/orders/:id/cancel", deps.Handlers.Order.Cancel)
				mutate.POST("/payments/init", deps.Handlers.Payment.Initiate)
			}

			adminMutate := authorized.Group("")
			adminMutate.Use(middleware.RequireAdmin(), middleware.Idempotency(deps.IdempStore))
			{
				adminMutate.POST("/vouchers/issue", deps.Handlers.Voucher.Issue)
				adminMutate.POST("/vouchers/:id/activate", deps.Handlers.Voucher.Activate)
				adminMutate.POST("/vouchers/:id/revoke", deps.Handlers.Voucher.Revoke)
				adminMutate.POST("/merchants", deps.Handlers.Merchant.Register)
				adminMutate.POST("/corporates", deps.Handlers.Corporate.Register)
				adminMutate.POST("/corporates/:id/budget", deps.Handlers.Corporate.SetBudget)
				adminMutate.POST("/distribution-jobs", deps.Handlers.Distribution.CreateJob)
				adminMutate.POST("/merchants/:id/stock/import", deps.Handlers.File.ImportStock)
				adminMutate.POST("/wallets/:ownerType/:ownerId/credit", deps.Handlers.Wallet.Credit)
			}

			authorized.GET("/vouchers/:id", deps.Handlers.Voucher.Get)
			authorized.GET("/vouchers/:id/validate", deps.Handlers.Voucher.Validate)
			authorized.GET("/vouchers", deps.Handlers.Voucher.ListMine)
			authorized.GET("/orders/:id", deps.Handlers.Order.Get)
			authorized.GET("/merchants/:id", deps.Handlers.Merchant.Get)
			authorized.GET("/merchants", deps.Handlers.Merchant.List)
			authorized.GET("/wallets/:ownerType/:ownerId/balance", deps.Handlers.Wallet.GetBalance)
			authorized.GET("/reports/redemptions", middleware.RequireAdmin(), deps.Handlers.Reporting.RedemptionRate)
			authorized.GET("/reports/corporates/:id/spend", middleware.RequireAdmin(), deps.Handlers.Reporting.CorporateSpend)
		}

		v1.POST("/claims/:token/claim", deps.Handlers.Distribution.Claim)
	}

	partner := r.Group("/partner/v1")
	partner.Use(middleware.PartnerAuth(deps.PartnerSvc, deps.RateLimiter))
	{
		partner.POST("/redeem", deps.Handlers.Partner.Redeem)
	}

	return r
}
