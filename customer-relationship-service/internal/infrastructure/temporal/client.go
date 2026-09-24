// Package temporal owns the Temporal client shared by the workflow worker and
// the orchestrator adapter.
package temporal

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

// Client wraps the SDK client. Client is nil when TEMPORAL_ADDRESS is not set.
type Client struct {
	Client client.Client
	Config config.TemporalConfig
}

func (c *Client) Enabled() bool { return c.Client != nil }

// New connects lazily: the service starts even if Temporal is not reachable
// yet, and the first call (or the worker's first poll) connects.
func New(lc fx.Lifecycle, cfg *config.Config) (*Client, error) {
	c := &Client{Config: cfg.Temporal}
	if cfg.Temporal.Address == "" {
		return c, nil
	}
	cl, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.Address,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal client: %w", err)
	}
	c.Client = cl
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		cl.Close()
		return nil
	}})
	return c, nil
}

var Module = fx.Options(fx.Provide(New))
