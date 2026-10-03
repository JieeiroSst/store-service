package ollama_model_py

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "ollama-model-py"

const DefaultBaseURL = "http://ollama-model-py:8000"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Tags(ctx context.Context) (*model.OllamaPyTags, error) {
	path := "/api/tags"
	q := url.Values{}
	h := http.Header{}
	var out *model.OllamaPyTags
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
