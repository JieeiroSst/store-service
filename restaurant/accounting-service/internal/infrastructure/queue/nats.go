package queue

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/config"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

// NewConn dials NATS directly rather than via logger.ConnectNats, whose
// deferred Close runs before the connection is ever returned to the caller.
func NewConn(cfg *config.Config) (*nats.Conn, error) {
	return nats.Connect(cfg.Nats.Dns)
}

func registerLifecycle(lc fx.Lifecycle, nc *nats.Conn) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			nc.Close()
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewConn),
	fx.Invoke(registerLifecycle),
)
