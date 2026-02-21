// Package di wires all application components using uber-go/fx.
// Follows the hexagonal architecture: core domain is injected with adapters via interfaces.
package di

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	apphttp "github.com/JIeeiroSst/wallet-service/internal/adapters/primary/http"
	"github.com/JIeeiroSst/wallet-service/internal/adapters/primary/http/handler"
	"github.com/JIeeiroSst/wallet-service/internal/adapters/secondary/postgres"
	"github.com/JIeeiroSst/wallet-service/internal/adapters/secondary/rediscache"
	"github.com/JIeeiroSst/wallet-service/internal/core/services"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/config"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/logger"
)

func provideLogger(cfg *config.Config) (*zap.Logger, error) {
	return logger.New(cfg.Log.Level)
}

var InfrastructureModule = fx.Module("infrastructure",
	fx.Provide(config.Load),
	fx.Provide(database.NewPostgresDB),
	fx.Provide(database.NewRedisClient),
	fx.Provide(provideLogger),
)

var RepositoryModule = fx.Module("repositories",
	fx.Provide(postgres.NewWalletRepository),
	fx.Provide(postgres.NewTransactionRepository),
	fx.Provide(postgres.NewCardRepository),
	fx.Provide(postgres.NewSettlementBatchRepository),
	fx.Provide(postgres.NewClearingRepository),
	fx.Provide(postgres.NewBankRepository),
	fx.Provide(postgres.NewMerchantRepository),
	fx.Provide(rediscache.NewCacheRepository),
)

var ServiceModule = fx.Module("services",
	fx.Provide(services.NewWalletService),
	fx.Provide(services.NewTransactionService),
	fx.Provide(services.NewCardService),
	fx.Provide(services.NewFeeCalculator),
)

var HandlerModule = fx.Module("handlers",
	fx.Provide(handler.NewWalletHandler),
	fx.Provide(handler.NewTransactionHandler),
	fx.Provide(func(wh *handler.WalletHandler, th *handler.TransactionHandler) *gin.Engine {
		return apphttp.NewRouter(wh, th)
	}),
)

var App = fx.Options(
	InfrastructureModule,
	RepositoryModule,
	ServiceModule,
	HandlerModule,
	fx.Invoke(startServer),
)

func startServer(lc fx.Lifecycle, engine *gin.Engine, cfg *config.Config, log *zap.Logger) {
	server := &http.Server{Addr: cfg.Server.Port, Handler: engine}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("wallet-service starting", zap.String("addr", cfg.Server.Port))
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("wallet-service shutting down")
			return server.Shutdown(ctx)
		},
	})
}
