package main

import (
	"log"

	"github.com/JIeeiroSst/solana-service/internal/adapters/blockchain"
	"github.com/JIeeiroSst/solana-service/internal/adapters/http"
	"github.com/JIeeiroSst/solana-service/internal/adapters/repository"
	"github.com/JIeeiroSst/solana-service/internal/adapters/wallet"
	"github.com/JIeeiroSst/solana-service/internal/config"
	"github.com/JIeeiroSst/solana-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Solana adapter https://solana.com
	solanaAdapter := blockchain.NewSolanaAdapter(cfg.SolanaRPCURL)

	// Initialize Circle adapters https://console.circle.com
	circleWalletAdapter := wallet.NewWalletAdapter(cfg.CircleAPIKey)
	circleTransactionAdapter := wallet.NewTransactionAdapter(cfg.CircleAPIKey)
	circleTokenAdapter := wallet.NewTokenAdapter(cfg.CircleAPIKey)

	// Initialize repositories
	txRepo := repository.NewMemoryTransactionRepository()
	accountRepo := repository.NewMemoryAccountRepository()
	bridgeRepo := repository.NewMemoryBridgeRepository()

	// Initialize Solana services
	accountService := services.NewAccountService(solanaAdapter, accountRepo)
	transactionService := services.NewTransactionService(solanaAdapter, txRepo)
	programService := services.NewProgramService(solanaAdapter)

	// Initialize Circle services
	circleWalletService := services.NewCircleWalletService(circleWalletAdapter)
	circleTokenService := services.NewCircleTokenService(circleTokenAdapter)
	circleTransactionService := services.NewCircleTransactionService(
		circleTransactionAdapter,
		solanaAdapter,
		bridgeRepo,
		transactionService,
	)

	// Initialize HTTP handlers
	handler := http.NewHandler(accountService, transactionService, programService)
	circleHandler := http.NewCircleHandler(circleWalletService, circleTransactionService, circleTokenService)

	// Setup router
	r := gin.Default()
	http.SetupRoutes(r, handler, circleHandler)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📡 Solana RPC: %s", cfg.SolanaRPCURL)
	log.Printf("🔵 Circle Wallets API: ENABLED")
	log.Println("✨ Full Circle integration ready!")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
