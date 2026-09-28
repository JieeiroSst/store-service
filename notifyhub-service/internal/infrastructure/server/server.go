package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/config"
	httpadapter "github.com/JIeeiroSst/notifyhub-service/internal/adapter/primary/http"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LC          fx.Lifecycle
	Shutdowner  fx.Shutdowner
	Handler     *httpadapter.Handler
	RateLimiter *httpadapter.RateLimiter
	Config      *config.Config
	Log         *zap.Logger
}

func New(p Params) {
	if p.Config.Server.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if p.Config.Server.APIKey == "" {
		p.Log.Warn("API_KEY is empty: /api/v1 is NOT authenticated")
	}

	engine := httpadapter.NewRouter(p.Handler, httpadapter.RouterConfig{
		APIKey:      p.Config.Server.APIKey,
		RateLimiter: p.RateLimiter,
		Log:         p.Log,
	})
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", p.Config.Server.Port),
		Handler:           engine,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", srv.Addr, err)
			}
			go func() {
				p.Log.Info("http server listening", zap.String("addr", srv.Addr))
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					p.Log.Error("http server error", zap.Error(err))
					_ = p.Shutdowner.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
