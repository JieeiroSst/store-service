package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/payment-wallet-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/payment-wallet-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/payment-wallet-service/internal/application"
	"github.com/JIeeiroSst/payment-wallet-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/payment-wallet-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.WalletRepository, port.TransactionRepository, port.TransferRepository, port.PaymentMethodRepository

	application.Module, // port.WalletUsecase, port.TransactionUsecase, port.PaymentMethodUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
