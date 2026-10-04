package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	dnsadapter "github.com/JIeeiroSst/networking-service/internal/adapter/primary/dns"
	httpadapter "github.com/JIeeiroSst/networking-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC          fx.Lifecycle
	Config      *config.Config
	Handler     *httpadapter.Handler
	DNS         *dnsadapter.Server
	Metrics     *metrics.Prometheus
	Catalog     *application.CatalogService
	Health      *application.HealthRunner
	Snapshotter *application.Snapshotter
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
			if err := p.Snapshotter.Restore(); err != nil {
				return fmt.Errorf("restore snapshot: %w", err)
			}
			return p.Catalog.RegisterSelf()
		},
		OnStop: func(context.Context) error {
			return p.Snapshotter.Stop()
		},
	})

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				log.Printf("networking-service HTTP API listening on %s", srv.Addr)
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

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error { return p.DNS.Start() },
		OnStop:  p.DNS.Stop,
	})

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.Health.Start()
			p.Snapshotter.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			p.Health.Stop()
			return nil
		},
	})
}
