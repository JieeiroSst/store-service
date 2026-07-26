package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/config"
	httpadapter "github.com/JIeeiroSst/admanagement-service/internal/adapter/primary/http"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Cfg     *config.Config
	Handler *httpadapter.Handler
}

func New(p Params) {
	router := mux.NewRouter()
	httpadapter.RegisterRoutes(router, p.Handler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Cfg.Server.Port),
		Handler: router,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				logrus.WithField("addr", srv.Addr).Info("admanagement-service: HTTP server listening")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logrus.WithError(err).Error("admanagement-service: HTTP server stopped")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
