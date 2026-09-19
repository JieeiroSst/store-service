package infrastructure

import (
	consumeradapter "github.com/JIeeiroSst/accounting-service/internal/adapter/primary/consumer"
	httpadapter "github.com/JIeeiroSst/accounting-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/accounting-service/internal/adapter/secondary/publisher"
	"github.com/JIeeiroSst/accounting-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/accounting-service/internal/application"
	"github.com/JIeeiroSst/accounting-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/accounting-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/accounting-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // *nats.Conn

	repository.Module, // port.AuthCartRepository
	publisher.Module,  // port.OrderPublisher

	application.Module, // port.AuthCartUsecase

	httpadapter.Module,     // *httpadapter.Handler
	consumeradapter.Module, // NATS subscriber lifecycle

	fx.Invoke(server.New),
)
