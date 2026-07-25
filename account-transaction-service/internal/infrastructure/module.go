package infrastructure

import (
	"github.com/Jieeirosst/account-transaction-service/internal/adapter/primary/grpc"
	"github.com/Jieeirosst/account-transaction-service/internal/adapter/secondary/repository"
	"github.com/Jieeirosst/account-transaction-service/internal/application"
	"github.com/Jieeirosst/account-transaction-service/internal/infrastructure/database"
	"github.com/Jieeirosst/account-transaction-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.AccountRepository, port.TransactionRepository

	application.Module, // port.AccountUsecase, port.TransactionUsecase

	grpc.Module, // *grpc.Handler

	fx.Invoke(server.New),
)
