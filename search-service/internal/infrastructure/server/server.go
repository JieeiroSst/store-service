package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/search-service/config"
	httpadapter "github.com/JIeeiroSst/search-service/internal/adapter/primary/http"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Config  *config.Config
	Handler *httpadapter.Handler
}

func New(p Params) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", p.Config.Server.ServerPort),
		Handler:           httpadapter.NewRouter(p.Handler),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      40 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				log.Printf("search-service %s HTTP API listening on %s", config.Version, srv.Addr)
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
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
