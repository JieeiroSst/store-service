package app

import (
	"os"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/delivery/http"
	"github.com/JIeeiroSst/upload-service/internal/repository"
	"github.com/JIeeiroSst/upload-service/internal/usecase"
	uploadAPI "github.com/JIeeiroSst/upload-service/pkg/api"
	"github.com/JIeeiroSst/upload-service/pkg/cache"
	"github.com/JIeeiroSst/upload-service/pkg/log"
	"github.com/JIeeiroSst/upload-service/pkg/mongo"
	"github.com/JIeeiroSst/upload-service/pkg/snowflake"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	config *config.Config
}

func NewServer(config *config.Config) *App {
	return &App{
		config: config,
	}
}

func (a *App) NewServerApp(router *fiber.App) {
	mongo, err := mongo.ConnectMongoDB(a.config.MongoDns)
	if err != nil {
		log.Error(err.Error())
	}

	snowflake := snowflake.NewSnowflake()
	uploadApi := uploadAPI.NewUploadFile(&uploadAPI.UploadApi{
		URL:   a.config.ImgBBApi,
		Token: a.config.TokenImgBB,
	})

	repository := repository.NewRepositories(mongo.Client)
	cache := cache.NewCacheHelper(a.config.HostCacheDNS)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repo:      *repository,
		Snowflake: snowflake,
		UploadApi: *uploadApi,
		Cache:     cache,
	})

	http := http.NewHandler(*usecase)
	http.Init(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = a.config.Port
	}
	router.Listen(port)
}
