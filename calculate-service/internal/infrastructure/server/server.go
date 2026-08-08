package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/calculate-service/config"
	httpadapter "github.com/JIeeiroSst/calculate-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Cfg     *config.Config
	Handler *httpadapter.Handler
}

// New builds the Gin HTTP server and ties its lifecycle to fx.
func New(p Params) {
	router := gin.Default()
	httpadapter.RegisterRoutes(router, p.Handler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", p.Cfg.Server.PortServer),
		Handler: router,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error(context.Background(), "http server error %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
