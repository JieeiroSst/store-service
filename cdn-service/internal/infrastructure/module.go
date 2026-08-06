package infrastructure

import (
	grpcadapter "github.com/JIeeiroSst/cdn-service/internal/adapter/primary/grpc"
	httpadapter "github.com/JIeeiroSst/cdn-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/cdn-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/cdn-service/internal/application"
	"github.com/JIeeiroSst/cdn-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/cdn-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *sql.DB

	repository.Module, // port.CDNRepository

	application.Module, // port.CDNUsecase

	grpcadapter.Module, // *grpcadapter.Handler
	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.NewGRPCServer),
	fx.Invoke(server.NewGatewayServer),
	fx.Invoke(server.NewHTTPServer),
)
