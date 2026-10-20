package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	httpadapter "github.com/JIeeiroSst/lucky-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/lucky-service/internal/config"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner
	Config     *config.Config
	Handler    *httpadapter.Handler
}

func New(p Params) {
	srv := &http.Server{
		Addr:              net.JoinHostPort("", strconv.Itoa(p.Config.HTTPPort)),
		Handler:           p.Handler.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				slog.Info("http server started", "addr", srv.Addr, "version", config.Version)
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("http server stopped", "err", err)
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
