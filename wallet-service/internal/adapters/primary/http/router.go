package http

import (
	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/wallet-service/internal/adapters/primary/http/handler"
	"github.com/JIeeiroSst/wallet-service/internal/adapters/primary/http/middleware"
)

func NewRouter(
	walletHandler *handler.WalletHandler,
	txHandler *handler.TransactionHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimiter(100, 50)) 

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "wallet-service"})
	})

	v1 := r.Group("/api/v1")

	// Wallet endpoints
	wallets := v1.Group("/wallets")
	{
		wallets.POST("", walletHandler.CreateWallet)
		wallets.GET("/:id", walletHandler.GetWallet)
		wallets.POST("/:id/credit", walletHandler.Credit)
		wallets.POST("/:id/freeze", walletHandler.FreezeWallet)
	}

	transactions := v1.Group("/transactions")
	{
		transactions.POST("/authorize", middleware.IdempotencyKeyRequired(), txHandler.Authorize)
		transactions.POST("/:id/capture", txHandler.Capture)
		transactions.POST("/:id/void", txHandler.Void)
		transactions.GET("/:id", txHandler.GetTransaction)
	}

	settlements := v1.Group("/settlements")
	{
		settlements.POST("/batches", txHandler.CreateSettlementBatch)
		settlements.POST("/batches/:batch_id/clear", txHandler.ProcessClearing)
		settlements.POST("/batches/:batch_id/settle", txHandler.ProcessSettlement)
	}

	return r
}
