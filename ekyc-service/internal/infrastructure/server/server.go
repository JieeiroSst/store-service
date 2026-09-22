package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.uber.org/fx"

	"github.com/JIeeiroSst/ekyc-service/config"
	httpadapter "github.com/JIeeiroSst/ekyc-service/internal/adapter/primary/http"
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
		Addr:    fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
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
