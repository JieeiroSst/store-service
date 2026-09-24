package infrastructure

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/upload-service/config"
	httpadapter "github.com/JIeeiroSst/upload-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/upload-service/internal/adapter/secondary/metadata"
	"github.com/JIeeiroSst/upload-service/internal/adapter/secondary/objectstore"
	"github.com/JIeeiroSst/upload-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/upload-service/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}

func validateConfig(cfg *config.Config) error { return cfg.Validate() }

func newMongo(lc fx.Lifecycle, cfg *config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: client.Disconnect})
	return client, nil
}

func newDatabase(c *mongo.Client, cfg *config.Config) *mongo.Database {
	return c.Database(cfg.Mongo.Database)
}

func serve(lc fx.Lifecycle, cfg *config.Config, engine *gin.Engine) {
	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: engine, ReadHeaderTimeout: 10 * time.Second}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logrus.WithError(err).Error("http server stopped")
				}
			}()
			logrus.Infof("listening on %s", srv.Addr)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	})
}

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(config.Load),
	fx.Invoke(validateConfig),
	fx.Provide(newMongo, newDatabase),
	metadata.Module,
	objectstore.Module,
	userservice.Module,
	application.Module,
	httpadapter.Module,
	fx.Invoke(serve),
)
