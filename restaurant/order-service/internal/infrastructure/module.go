package infrastructure

import (
	consumeradapter "github.com/JIeeiroSst/order-service/internal/adapter/primary/consumer"
	httpadapter "github.com/JIeeiroSst/order-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/order-service/internal/adapter/secondary/proxy"
	"github.com/JIeeiroSst/order-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/order-service/internal/application"
	"github.com/JIeeiroSst/order-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/order-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/order-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // *nats.Conn

	repository.Module, // port.{Order,Reservation}Repository
	proxy.Module,      // port.FoodPricer (kitchen-service HTTP client)

	application.Module, // port.{Order,Reservation}Usecase

	httpadapter.Module,     // *httpadapter.Handler
	consumeradapter.Module, // NATS subscriber lifecycle

	fx.Invoke(server.New),
)
