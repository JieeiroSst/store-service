package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	httpadapter "github.com/JIeeiroSst/webrtc-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/metrics"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	Metrics *metrics.Prometheus
	Config  *config.Config
}

func New(p Params) {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", p.Metrics.Handler())
	mux.Handle("/", httpadapter.NewRouter(p.Handler))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {

			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				log.Printf("webrtc-service listening on %s", srv.Addr)
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := srv.Shutdown(ctx)
			p.Handler.Shutdown()
			return err
		},
	})
}
