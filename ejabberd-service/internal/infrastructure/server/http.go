package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	httpapi "github.com/JIeeiroSst/ejabberd-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Config  *config.Config
	Handler *httpapi.Handler
	TLS     *tls.Config `optional:"true"`
}

func NewHTTP(p Params) error {
	var servers []*http.Server
	for _, l := range p.Config.Listen {
		if l.Module != "ejabberd_http" {
			if l.Module != "ejabberd_c2s" {
				slog.Warn("unsupported listener module, skipping", "port", l.Port, "module", l.Module)
			}
			continue
		}
		if l.TLS && p.TLS == nil {
			return fmt.Errorf("http listener on port %d requires tls but no certificate is available", l.Port)
		}
		ip := l.IP
		if ip == "::" || ip == "0.0.0.0" {
			ip = ""
		}
		srv := &http.Server{
			Addr:              net.JoinHostPort(ip, strconv.Itoa(l.Port)),
			Handler:           p.Handler.Router(l),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      40 * time.Second,
			IdleTimeout:       2 * time.Minute,
		}
		useTLS := l.TLS
		if useTLS {
			srv.TLSConfig = p.TLS
		}
		servers = append(servers, srv)
		p.LC.Append(fx.Hook{
			OnStart: func(context.Context) error {
				ln, err := net.Listen("tcp", srv.Addr)
				if err != nil {
					return err
				}
				if useTLS {
					ln = tls.NewListener(ln, srv.TLSConfig)
				}
				go func() {
					slog.Info("http listener started", "addr", srv.Addr, "tls", useTLS, "version", config.Version)
					if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
						slog.Error("http server error", "addr", srv.Addr, "err", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
	}
	if len(servers) == 0 {
		slog.Warn("no ejabberd_http listener configured; health endpoints are unavailable")
	}
	return nil
}
