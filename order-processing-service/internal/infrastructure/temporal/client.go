package temporal

import (
	"fmt"

	"github.com/JIeeiroSst/order-processing-service/internal/config"
	"go.temporal.io/sdk/client"
)

// NewClient creates a new Temporal client using the provided configuration.
// The client is used to:
//   - Start Workflow Executions
//   - Signal Workflow Executions
//   - Query Workflow Executions
//   - Terminate or Cancel Workflow Executions
//   - Complete Activities asynchronously
func NewClient(cfg config.TemporalConfig) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}
	return c, nil
}
