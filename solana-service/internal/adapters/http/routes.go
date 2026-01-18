package http

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, h *Handler, ch *CircleHandler) {
	api := r.Group("/api/v1")
	{
		// Solana routes
		solana := api.Group("/solana")
		{
			// Accounts
			solana.GET("/accounts/:address", h.GetAccount)
			solana.GET("/accounts/:address/balance", h.GetBalance)

			// Transactions
			solana.GET("/transactions/:signature", h.GetTransaction)
			solana.POST("/transactions/estimate-fee", h.EstimateFee)

			// Programs
			solana.GET("/programs/:id", h.GetProgram)
			solana.POST("/programs/pda", h.FindPDA)
		}

		// Circle routes
		circle := api.Group("/circle")
		{
			// Wallet Sets
			circle.POST("/wallet-sets", ch.CreateWalletSet)
			circle.GET("/wallet-sets/:id", ch.GetWalletSet)
			circle.GET("/wallet-sets", ch.ListWalletSets)

			// Wallets
			circle.POST("/wallets", ch.CreateWallet)
			circle.GET("/wallets/:id", ch.GetWallet)
			circle.GET("/wallets", ch.ListWallets)
			circle.GET("/wallets/:id/balance", ch.GetWalletBalance)
			circle.GET("/wallets/:id/nfts", ch.GetWalletNFTs)
			circle.PUT("/wallets/:id/freeze", ch.FreezeWallet)
			circle.PUT("/wallets/:id/unfreeze", ch.UnfreezeWallet)

			// Transactions
			circle.POST("/transactions/transfer", ch.CreateTransfer)
			circle.POST("/transactions/nft-transfer", ch.CreateNFTTransfer)
			circle.POST("/transactions/contract-execution", ch.ExecuteContract)
			circle.GET("/transactions/:id", ch.GetTransaction)
			circle.GET("/transactions/estimate-fee", ch.EstimateFee)

			// Tokens
			circle.GET("/tokens/:id", ch.GetToken)
			circle.GET("/tokens", ch.ListTokens)
			circle.POST("/tokens/import", ch.ImportToken)
		}

		// Bridge routes
		bridge := api.Group("/bridge")
		{
			bridge.POST("/solana-to-circle", ch.BridgeSolanaToCircle)
			bridge.POST("/circle-to-solana", ch.BridgeCircleToSolana)
			bridge.GET("/transfers/:id", ch.GetBridgeTransfer)
		}
	}
}
