package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/automatic-payment-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/automatic-payment-service/internal/adapter/secondary/gateway"
	"github.com/JIeeiroSst/automatic-payment-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/automatic-payment-service/internal/application"
	"github.com/JIeeiroSst/automatic-payment-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/automatic-payment-service/internal/infrastructure/scheduler"
	"github.com/JIeeiroSst/automatic-payment-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	gateway.Module,    // port.PaymentGatewayPort
	repository.Module, // port.*Repository

	application.Module, // port.*Usecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
	fx.Invoke(scheduler.New),
)
