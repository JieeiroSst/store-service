package ai_agent_system

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "ai-agent-system"

const DefaultBaseURL = "http://ai-agent-system:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) ChatHistory(ctx context.Context, limit *int, xUserID string) (*model.AiAgentHistory, error) {
	path := "/v1/chat/history"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	h.Set("X-User-Id", xUserID)
	var out *model.AiAgentHistory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Tags(ctx context.Context) (*model.AiAgentTags, error) {
	path := "/api/tags"
	q := url.Values{}
	h := http.Header{}
	var out *model.AiAgentTags
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
