package nodecall

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

// httpNodeCaller lets the edge role reach a specific transcode node's
// internal HTTP surface - currently just force-unpublish, called from
// ModerationUsecase.ForceEndStream. Cluster-internal only: authenticated
// with a shared secret, not a user JWT (see middleware.RequireInternalSecret
// on the node role's receiving end).
type httpNodeCaller struct {
	client         *http.Client
	internalSecret string
}

func NewNodeCaller(cfg *config.Config) port.NodeCaller {
	return &httpNodeCaller{
		client:         &http.Client{Timeout: 10 * time.Second},
		internalSecret: cfg.Internal.SharedSecret,
	}
}

func (c *httpNodeCaller) ForceUnpublish(ctx context.Context, nodeHTTPAddr, streamKey string) error {
	url := fmt.Sprintf("%s/internal/streams/%s/force-unpublish", nodeHTTPAddr, streamKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("build force-unpublish request: %w", err)
	}
	req.Header.Set("X-Internal-Token", c.internalSecret)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call node %s: %w", nodeHTTPAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("node %s rejected force-unpublish: status %d", nodeHTTPAddr, resp.StatusCode)
	}
	return nil
}
