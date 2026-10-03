package upload_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "upload_service"

const DefaultBaseURL = "http://upload-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Files(ctx context.Context, receiverID *string, limit *int, offset *int) (*model.UploadFileList, error) {
	path := "/api/v1/upload"
	q := url.Values{}
	if receiverID != nil {
		q.Set("receiver_id", *receiverID)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.UploadFileList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) File(ctx context.Context, id string) (*model.UploadFile, error) {
	path := "/api/v1/upload/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.UploadFile
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
