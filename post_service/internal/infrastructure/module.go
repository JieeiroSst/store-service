package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/post-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/post-service/internal/adapter/secondary/idgen"
	"github.com/JIeeiroSst/post-service/internal/adapter/secondary/objectstorage"
	"github.com/JIeeiroSst/post-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/post-service/internal/application"
	"github.com/JIeeiroSst/post-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/post-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module,    // port.PostRepository, port.CategoryRepository, port.MediaRepository
	objectstorage.Module, // port.ObjectStorage
	idgen.Module,         // port.IDGenerator

	application.Module, // port.PostUsecase, port.CategoryUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
