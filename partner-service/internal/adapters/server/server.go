package server

import (
	"context"
	"net/http"

	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC     fx.Lifecycle
	Router *gin.Engine
	Config *config.Config
}

func New(p Params) {
	srv := &http.Server{
		Addr:    p.Config.Server.ServerPort,
		Handler: p.Router,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Log.Errorf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
