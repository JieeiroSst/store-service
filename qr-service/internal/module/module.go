package module

import (
	"github.com/JIeeiroSst/qr-service/internal/application/service"
	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/config"
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/http/handler"
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/http/router"
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/repository"
	"github.com/JIeeiroSst/qr-service/pkg/logger"
	"go.uber.org/fx"
)

var ConfigModule = fx.Module("config",
	fx.Provide(config.Load),
)

var LoggerModule = fx.Module("logger",
	fx.Provide(logger.NewLogger),
)

var DatabaseModule = fx.Module("database",
	fx.Provide(repository.NewMongoDatabase),
)

var RepositoryModule = fx.Module("repository",
	fx.Provide(
		fx.Annotate(
			repository.NewMongoQRCodeRepository,
			fx.As(new(port.QRCodeRepository)),
		),
		fx.Annotate(
			repository.NewMongoScanHistoryRepository,
			fx.As(new(port.ScanHistoryRepository)),
		),
	),
)

var ServiceModule = fx.Module("service",
	fx.Provide(
		fx.Annotate(
			service.NewQRCodeService,
			fx.As(new(port.QRCodeService)),
		),
		fx.Annotate(
			service.NewScanHistoryService,
			fx.As(new(port.ScanHistoryService)),
		),
	),
)

var HandlerModule = fx.Module("handler",
	fx.Provide(handler.NewQRHandler),
)

var RouterModule = fx.Module("router",
	fx.Provide(router.NewRouter),
)

var AllModules = fx.Options(
	ConfigModule,
	LoggerModule,
	DatabaseModule,
	RepositoryModule,
	ServiceModule,
	HandlerModule,
	RouterModule,
)
