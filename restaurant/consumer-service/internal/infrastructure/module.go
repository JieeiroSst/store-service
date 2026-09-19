package infrastructure

import (
	consumeradapter "github.com/JIeeiroSst/consumer-service/internal/adapter/primary/consumer"
	httpadapter "github.com/JIeeiroSst/consumer-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/consumer-service/internal/adapter/secondary/publisher"
	"github.com/JIeeiroSst/consumer-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/consumer-service/internal/application"
	"github.com/JIeeiroSst/consumer-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/consumer-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/consumer-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // *nats.Conn

	repository.Module, // port.ConsumerRepository
	publisher.Module,  // port.OrderPublisher

	application.Module, // port.ConsumerUsecase

	httpadapter.Module,     // *httpadapter.Handler
	consumeradapter.Module, // NATS subscriber lifecycle

	fx.Invoke(server.New),
)
