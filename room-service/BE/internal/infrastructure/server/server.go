package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, engine *gin.Engine, cfg *config.Config) {
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
}
