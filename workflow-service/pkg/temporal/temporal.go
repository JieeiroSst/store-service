package temporal

import (
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"go.temporal.io/sdk/client"
)

func NewWorkflow(host string) client.Client {
	temporal, err := client.NewClient(client.Options{
		HostPort: host,
	})
	if err != nil {
		log.Error("unable to create Temporal client", err)
	}
	return temporal
}