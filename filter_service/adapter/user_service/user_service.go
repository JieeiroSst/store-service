package user_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "user_service"

const DefaultBaseURL = "http://user-api-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) FindUser(ctx context.Context, username *string, email *string, page *int, limit *int) (*model.UserFindUserResponse, error) {
	path := "/user"
	q := url.Values{}
	if username != nil {
		q.Set("username", *username)
	}
	if email != nil {
		q.Set("email", *email)
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.UserFindUserResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetRole(ctx context.Context, id int) (*model.UserGetRoleResponse, error) {
	path := "/api/v1/role/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.UserGetRoleResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListRoles(ctx context.Context, limit *int, page *int) (*model.UserListRolesResponse, error) {
	path := "/api/v1/role"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	h := http.Header{}
	var out *model.UserListRolesResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
