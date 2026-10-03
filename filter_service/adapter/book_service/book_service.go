package book_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "book-service"

const DefaultBaseURL = "http://book-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Books(ctx context.Context, cursor *string, limit *int, qArg *string, author *string, year *int, category *string) (*model.BookSvcBookPage, error) {
	path := "/api/v1/books"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if author != nil {
		q.Set("author", *author)
	}
	if year != nil {
		q.Set("year", strconv.Itoa(*year))
	}
	if category != nil {
		q.Set("category", *category)
	}
	h := http.Header{}
	var out *model.BookSvcBookPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Book(ctx context.Context, id int) (*model.BookSvcBookDetail, error) {
	path := "/api/v1/books/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookSvcBookDetail
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Chapters(ctx context.Context, id int, cursor *string, limit *int) (*model.BookSvcChapterPage, error) {
	path := "/api/v1/books/" + url.PathEscape(strconv.Itoa(id)) + "/chapters"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.BookSvcChapterPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Chapter(ctx context.Context, id int, number int) (*model.BookSvcChapter, error) {
	path := "/api/v1/books/" + url.PathEscape(strconv.Itoa(id)) + "/chapters/" + url.PathEscape(strconv.Itoa(number))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookSvcChapter
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context) ([]*model.BookSvcCategory, error) {
	path := "/api/v1/categories"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookSvcCategory
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Category(ctx context.Context, slug string) (*model.BookSvcCategory, error) {
	path := "/api/v1/categories/" + url.PathEscape(slug)
	q := url.Values{}
	h := http.Header{}
	var out *model.BookSvcCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CategoryBooks(ctx context.Context, slug string, cursor *string, limit *int, qArg *string) (*model.BookSvcBookPage, error) {
	path := "/api/v1/categories/" + url.PathEscape(slug) + "/books"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	h := http.Header{}
	var out *model.BookSvcBookPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
