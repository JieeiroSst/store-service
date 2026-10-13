package http

import (
	"net/http"

	"github.com/JIeeiroSst/payment-wallet-service/pkg/authn"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1", authenticate(authn.FromEnv()), requireUser())
	{
		api.POST("/wallets", forceUserBody("user_id"), h.CreateWallet)
		api.GET("/wallets/:id", h.walletParam("id"), h.GetWallet)
		api.GET("/wallets/user/:userId", userParam("userId"), h.GetWalletByUser)
		api.POST("/wallets/:id/deposit", requireAdmin(), h.Deposit)
		api.POST("/wallets/:id/withdraw", requireAdmin(), h.Withdraw)
		api.GET("/wallets/:id/transactions", h.walletParam("id"), h.ListTransactions)
		api.GET("/wallets/:id/transactions/by-reference/:referenceId", h.walletParam("id"), h.GetTransactionByReference)
		api.GET("/wallets/:id/statement", h.walletParam("id"), h.Statement)
		api.POST("/wallets/:id/freeze", requireAdmin(), h.FreezeWallet)
		api.POST("/wallets/:id/unfreeze", requireAdmin(), h.UnfreezeWallet)
		api.POST("/wallets/:id/close", requireAdmin(), h.CloseWallet)
		api.PATCH("/wallets/:id/limits", requireAdmin(), h.SetLimits)

		api.POST("/wallets/:id/pockets", h.walletParam("id"), h.CreatePocket)
		api.GET("/wallets/:id/pockets", h.walletParam("id"), h.ListPockets)
		api.POST("/pockets/:pocketId/deposit", h.pocketParam, h.DepositToPocket)
		api.POST("/pockets/:pocketId/withdraw", h.pocketParam, h.WithdrawFromPocket)
		api.DELETE("/pockets/:pocketId", h.pocketParam, h.ClosePocket)

		api.POST("/transfers", h.walletBody("sender_wallet_id", true), h.Transfer)
		api.GET("/transfers/:id", h.transferParam, h.GetTransfer)
		api.POST("/transfers/:id/reverse", requireAdmin(), h.ReverseTransfer)
		api.GET("/transactions/:id", h.transactionParam, h.GetTransaction)
		api.POST("/transactions/:id/reverse", requireAdmin(), h.ReverseTransaction)

		api.POST("/payment-requests", h.walletBody("requester_wallet_id", true), h.CreatePaymentRequest)
		api.GET("/payment-requests", h.walletQuery("wallet_id"), h.ListPaymentRequests)
		api.GET("/payment-requests/:id", h.GetPaymentRequest)
		api.GET("/payment-requests/:id/qr", h.GetPaymentRequestQR)
		api.POST("/payment-requests/:id/pay", h.walletBody("payer_wallet_id", true), h.PayPaymentRequest)
		api.POST("/payment-requests/:id/cancel", h.requesterParam, h.CancelPaymentRequest)

		api.POST("/payment-methods", forceUserBody("user_id"), h.AddPaymentMethod)
		api.GET("/payment-methods", userQuery("user_id"), h.ListPaymentMethods)
		api.DELETE("/payment-methods/:id", h.paymentMethodParam, h.RemovePaymentMethod)
		api.PATCH("/payment-methods/:id/default", h.paymentMethodParam, h.SetDefaultPaymentMethod)
	}

	return engine
}
