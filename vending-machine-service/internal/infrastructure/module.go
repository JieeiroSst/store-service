package infrastructure

import (
	"github.com/JIeeiroSst/vending-machine-service/config"
	httpadapter "github.com/JIeeiroSst/vending-machine-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/primary/worker"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/coupon"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/notification"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/payment"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/postgres"
	"github.com/JIeeiroSst/vending-machine-service/internal/application"
	"github.com/JIeeiroSst/vending-machine-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/vending-machine-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(config.Load),

	database.Module,
	postgres.Module,
	payment.Module,
	coupon.Module,
	notification.Module,

	application.Module,

	httpadapter.Module,
	worker.Module,

	fx.Invoke(server.New),
)
