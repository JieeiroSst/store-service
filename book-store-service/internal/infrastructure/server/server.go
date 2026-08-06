package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/JIeeiroSst/bookStore-service/config"
	httpadapter "github.com/JIeeiroSst/bookStore-service/internal/adapter/primary/http"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	Config  *config.Config
}

func New(p Params) {
	engine := httpadapter.NewRouter(p.Handler)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Config.Server.ServerPort),
		Handler: engine,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
