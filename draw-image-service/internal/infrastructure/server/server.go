package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JIeeiroSst/draw-image-service/config"
	httpadapter "github.com/JIeeiroSst/draw-image-service/internal/adapter/primary/http"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	Config  *config.Config
}

func New(p Params) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
		Handler:           httpadapter.NewRouter(p.Handler),
		ReadHeaderTimeout: 10 * time.Second,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				log.Printf("draw-image-service listening on %s", srv.Addr)
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
