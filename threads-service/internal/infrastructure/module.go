package infrastructure

import (
	"github.com/JIeeiroSst/threads-service/config"
	httpadapter "github.com/JIeeiroSst/threads-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/threads-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/threads-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/threads-service/internal/adapter/secondary/trending"
	"github.com/JIeeiroSst/threads-service/internal/adapter/secondary/userclient"
	"github.com/JIeeiroSst/threads-service/internal/application"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/JIeeiroSst/threads-service/internal/infrastructure/database"
	redisinfra "github.com/JIeeiroSst/threads-service/internal/infrastructure/redis"
	"github.com/JIeeiroSst/threads-service/internal/infrastructure/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func newPostRepository(db *gorm.DB, redisClient *redis.Client) port.PostRepository {
	return cache.NewCachedPostRepository(repository.NewPostRepository(db), redisClient)
}

func newUserClient(cfg *config.Config, redisClient *redis.Client) port.UserClient {
	return cache.NewCachedUserClient(userclient.NewUserClient(cfg), redisClient)
}

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module,   // *gorm.DB
	redisinfra.Module, // *redis.Client

	fx.Provide(repository.NewCommentRepository),
	fx.Provide(repository.NewLikeRepository),
	fx.Provide(repository.NewFollowRepository),
	fx.Provide(repository.NewTagRepository),
	fx.Provide(repository.NewBookmarkRepository),
	fx.Provide(newPostRepository), // port.PostRepository, Redis-cached (see above)
	fx.Provide(newUserClient),     // port.UserClient, Redis-cached (see above)

	trending.Module, // port.TrendingTagsStore

	application.Module, // port.PostUsecase, port.CommentUsecase, port.LikeUsecase, port.FollowUsecase, port.BookmarkUsecase, port.TagUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
