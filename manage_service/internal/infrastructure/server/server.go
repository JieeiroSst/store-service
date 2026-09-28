package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/manage-service/config"
	httpadapter "github.com/JIeeiroSst/manage-service/internal/adapter/primary/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner
	Handler    *httpadapter.Handler
	Config     *config.Config
	Log        *zap.Logger
}

func New(p Params) {
	if p.Config.Secret.AuthorizeKey == "" {
		p.Log.Warn("AUTHORIZE_KEY is empty: API is not protected by X-Api-Key")
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", p.Config.Server.ServerPort),
		Handler: httpadapter.NewRouter(p.Handler, httpadapter.RouterConfig{
			AuthorizeKey: p.Config.Secret.AuthorizeKey,
			Timeout:      60 * time.Second,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      65 * time.Second,
		IdleTimeout:       60 * time.Second,
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
