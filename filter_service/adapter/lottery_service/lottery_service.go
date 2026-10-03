package lottery_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "lottery-service"

const DefaultBaseURL = "http://lottery-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Results(ctx context.Context, date string, province string) (*model.LotteryResult, error) {
	path := "/api/v1/results"
	q := url.Values{}
	q.Set("date", date)
	q.Set("province", province)
	h := http.Header{}
	var out *model.LotteryResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
