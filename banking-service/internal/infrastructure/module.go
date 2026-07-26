package infrastructure

import (
	httpadapter "github.com/JieeiroSst/banking-service/internal/adapter/primary/http"
	kafkaadapter "github.com/JieeiroSst/banking-service/internal/adapter/primary/kafka"
	"github.com/JieeiroSst/banking-service/internal/adapter/secondary/repository"
	"github.com/JieeiroSst/banking-service/internal/application"
	"github.com/JieeiroSst/banking-service/internal/infrastructure/cache"
	"github.com/JieeiroSst/banking-service/internal/infrastructure/database"
	"github.com/JieeiroSst/banking-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	cache.Module, // *redis.Client

	repository.Module, // port.*Repository

	application.Module, // port.*Usecase

	httpadapter.Module, // *httpadapter.Handler

	kafkaadapter.Module, // consumes the transaction topic

	fx.Invoke(server.New),
)
