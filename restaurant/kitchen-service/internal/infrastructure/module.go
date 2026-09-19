package infrastructure

import (
	consumeradapter "github.com/JIeeiroSst/kitchen-service/internal/adapter/primary/consumer"
	httpadapter "github.com/JIeeiroSst/kitchen-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/kitchen-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/kitchen-service/internal/application"
	"github.com/JIeeiroSst/kitchen-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/kitchen-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/kitchen-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // *nats.Conn

	repository.Module, // port.{Kitchen,Food,Category}Repository

	application.Module, // port.{Kitchen,Food,Category}Usecase

	httpadapter.Module,     // *httpadapter.Handler
	consumeradapter.Module, // NATS subscriber lifecycle

	fx.Invoke(server.New),
)
