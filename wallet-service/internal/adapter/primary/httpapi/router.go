package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type Services struct {
	Wallets         port.WalletService
	Funding         port.FundingService
	Transfers       port.TransferService
	Payments        port.PaymentService
	CardPayments    port.CardPaymentService
	Settlements     port.SettlementService
	History         port.HistoryService
	Cards           port.CardService
	BankAccounts    port.BankAccountService
	PaymentRequests port.PaymentRequestService
}

type API struct {
	Services
	logger *zap.Logger
}

type RouterOptions struct {
	RateLimitCapacity float64
	RateLimitPerSec   float64
	Ready             func(ctx context.Context) error
}

func NewRouter(svc Services, opts RouterOptions, logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	a := &API{Services: svc, logger: logger}

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "wallet-service"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		if opts.Ready != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := opts.Ready(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	if opts.RateLimitCapacity > 0 && opts.RateLimitPerSec > 0 {
		r.Use(rateLimiter(opts.RateLimitCapacity, opts.RateLimitPerSec))
	}

	v1 := r.Group("/api/v1")

	wallets := v1.Group("/wallets")
	wallets.POST("", a.createWallet)
	wallets.GET("/:id", a.getWallet)
	wallets.GET("/:id/balance", a.getBalance)
	wallets.GET("/:id/limits", a.getLimits)
	wallets.POST("/:id/credit", a.credit)
	wallets.POST("/:id/freeze", a.freezeWallet)
	wallets.POST("/:id/unfreeze", a.unfreezeWallet)
	wallets.POST("/:id/close", a.closeWallet)
	wallets.PUT("/:id/tier", a.changeTier)
	wallets.PUT("/:id/pin", a.setPin)
	wallets.POST("/:id/topup", a.topUp)
	wallets.POST("/:id/withdraw", a.withdraw)
	wallets.GET("/:id/transactions", a.listTransactions)
	wallets.GET("/:id/ledger", a.listLedger)
	wallets.GET("/:id/statement", a.statement)
	wallets.POST("/:id/cards", a.issueCard)
	wallets.GET("/:id/cards", a.listCards)

	users := v1.Group("/users/:user_id")
	users.GET("/wallets", a.listUserWallets)
	users.GET("/wallets/:currency", a.getUserWallet)
	users.POST("/bank-accounts", a.linkBankAccount)
	users.GET("/bank-accounts", a.listBankAccounts)
	users.DELETE("/bank-accounts/:account_id", a.removeBankAccount)
	users.POST("/bank-accounts/:account_id/default", a.setDefaultBankAccount)
	users.GET("/payment-requests", a.listPaymentRequests)

	v1.POST("/transfers", a.transfer)
	v1.POST("/payments", a.pay)

	transactions := v1.Group("/transactions")
	transactions.POST("/authorize", a.authorize)
	transactions.GET("/:id", a.getTransaction)
	transactions.POST("/:id/capture", a.capture)
	transactions.POST("/:id/void", a.void)
	transactions.POST("/:id/refund", a.refund)

	cards := v1.Group("/cards")
	cards.GET("/:id", a.getCard)
	cards.POST("/:id/block", a.blockCard)
	cards.POST("/:id/unblock", a.unblockCard)

	requests := v1.Group("/payment-requests")
	requests.POST("", a.createPaymentRequest)
	requests.GET("/:id", a.getPaymentRequest)
	requests.POST("/:id/pay", a.payPaymentRequest)
	requests.POST("/:id/decline", a.declinePaymentRequest)
	requests.POST("/:id/cancel", a.cancelPaymentRequest)

	settlements := v1.Group("/settlements")
	settlements.POST("/batches", a.createSettlementBatch)
	settlements.POST("/batches/:batch_id/clear", a.clearBatch)
	settlements.POST("/batches/:batch_id/settle", a.settleBatch)

	return r
}
