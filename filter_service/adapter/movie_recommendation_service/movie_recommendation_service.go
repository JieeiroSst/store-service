package movie_recommendation_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "movie-recommendation-service"

const DefaultBaseURL = "http://movie-recommendation-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) ForUser(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/users/" + url.PathEscape(userID)
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Home(ctx context.Context, userID string) (*model.MovieHome, error) {
	path := "/api/recommendations/users/" + url.PathEscape(userID) + "/home"
	q := url.Values{}
	h := http.Header{}
	var out *model.MovieHome
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContinueWatching(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/users/" + url.PathEscape(userID) + "/continue-watching"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) History(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/users/" + url.PathEscape(userID) + "/history"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) NewReleases(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/new"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Similar(ctx context.Context, id string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/videos/" + url.PathEscape(id) + "/similar"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Trending(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	path := "/api/recommendations/trending"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if snapshot != nil {
		q.Set("snapshot", *snapshot)
	}
	h := http.Header{}
	var out *model.MoviePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
