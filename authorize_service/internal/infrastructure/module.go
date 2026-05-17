package infrastructure

import (
	"github.com/JieeiroSst/authorize-service/internal/adapter/primary/grpc"
	"github.com/JieeiroSst/authorize-service/internal/adapter/secondary/cache"
	"github.com/JieeiroSst/authorize-service/internal/adapter/secondary/otp"
	"github.com/JieeiroSst/authorize-service/internal/adapter/secondary/repository"
	"github.com/JieeiroSst/authorize-service/internal/application"
	"github.com/JieeiroSst/authorize-service/internal/infrastructure/database"
	"github.com/JieeiroSst/authorize-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newConfig),

	database.Module, // *gorm.DB, persist.Adapter

	cache.Module,      // port.CachePort
	otp.Module,        // port.OTPPort
	repository.Module, // port.CasbinRepository

	application.Module, // port.CasbinUsecase, port.OTPUsecase

	grpc.Module, // *grpc.Handler

	fx.Invoke(server.New),
)
