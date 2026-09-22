package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1")
	{
		api.POST("/wallets", h.CreateWallet)
		api.GET("/wallets/:id", h.GetWallet)
		api.GET("/wallets/user/:userId", h.GetWalletByUser)
		api.POST("/wallets/:id/deposit", h.Deposit)
		api.POST("/wallets/:id/withdraw", h.Withdraw)
		api.GET("/wallets/:id/transactions", h.ListTransactions)
		api.GET("/wallets/:id/statement", h.Statement)
		api.POST("/wallets/:id/freeze", h.FreezeWallet)
		api.POST("/wallets/:id/unfreeze", h.UnfreezeWallet)
		api.POST("/wallets/:id/close", h.CloseWallet)
		api.PATCH("/wallets/:id/limits", h.SetLimits)

		api.POST("/wallets/:id/pockets", h.CreatePocket)
		api.GET("/wallets/:id/pockets", h.ListPockets)
		api.POST("/pockets/:pocketId/deposit", h.DepositToPocket)
		api.POST("/pockets/:pocketId/withdraw", h.WithdrawFromPocket)
		api.DELETE("/pockets/:pocketId", h.ClosePocket)

		api.POST("/transfers", h.Transfer)
		api.GET("/transfers/:id", h.GetTransfer)
		api.POST("/transfers/:id/reverse", h.ReverseTransfer)
		api.GET("/transactions/:id", h.GetTransaction)
		api.POST("/transactions/:id/reverse", h.ReverseTransaction)

		api.POST("/payment-requests", h.CreatePaymentRequest)
		api.GET("/payment-requests", h.ListPaymentRequests)
		api.GET("/payment-requests/:id", h.GetPaymentRequest)
		api.GET("/payment-requests/:id/qr", h.GetPaymentRequestQR)
		api.POST("/payment-requests/:id/pay", h.PayPaymentRequest)
		api.POST("/payment-requests/:id/cancel", h.CancelPaymentRequest)

		api.POST("/payment-methods", h.AddPaymentMethod)
		api.GET("/payment-methods", h.ListPaymentMethods)
		api.DELETE("/payment-methods/:id", h.RemovePaymentMethod)
		api.PATCH("/payment-methods/:id/default", h.SetDefaultPaymentMethod)
	}

	return engine
}
