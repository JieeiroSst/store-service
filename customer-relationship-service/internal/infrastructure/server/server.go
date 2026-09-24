package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, cfg *config.Config, engine *gin.Engine) {
	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: engine}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logrus.WithError(err).Error("http server stopped")
				}
			}()
			logrus.Infof("listening on %s", srv.Addr)
			return nil
		},
		OnStop: srv.Shutdown,
	})
}
