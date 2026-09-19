package infrastructure

import (
	consumeradapter "github.com/JIeeiroSst/delivery-service/internal/adapter/primary/consumer"
	httpadapter "github.com/JIeeiroSst/delivery-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/delivery-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/delivery-service/internal/application"
	"github.com/JIeeiroSst/delivery-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/delivery-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/delivery-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // *nats.Conn

	repository.Module, // port.DeliveryRepository

	application.Module, // port.DeliveryUsecase

	httpadapter.Module,     // *httpadapter.Handler
	consumeradapter.Module, // NATS subscriber lifecycle

	fx.Invoke(server.New),
)
