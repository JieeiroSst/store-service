package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	httpadapter "github.com/JIeeiroSst/auth-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/metrics"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Config  *config.Config
	Handler *httpadapter.Handler
	Metrics *metrics.Prometheus
}

func New(p Params) {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", p.Metrics.Handler())
	mux.Handle("/", httpadapter.NewRouter(p.Handler))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
		Handler:           http.TimeoutHandler(mux, 30*time.Second, `{"error":"request timed out"}`),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      35 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				log.Printf("auth-service HTTP API listening on %s", srv.Addr)
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
