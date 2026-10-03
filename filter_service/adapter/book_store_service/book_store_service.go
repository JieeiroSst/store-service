package book_store_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "book-store-service"

const DefaultBaseURL = "http://book-store-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Books(ctx context.Context) ([]*model.BookStoreBook, error) {
	path := "/api/v1/books"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookStoreBook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) HighestPriceBook(ctx context.Context) (*model.BookStoreBook, error) {
	path := "/api/v1/books/highest-price"
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreBook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MostPurchasedBook(ctx context.Context) (*model.BookStoreBook, error) {
	path := "/api/v1/books/most-purchased"
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreBook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Book(ctx context.Context, id int) (*model.BookStoreBook, error) {
	path := "/api/v1/books/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreBook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Authors(ctx context.Context) ([]*model.BookStoreAuthor, error) {
	path := "/api/v1/authors"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookStoreAuthor
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MostReadAuthor(ctx context.Context) (*model.BookStoreAuthorStat, error) {
	path := "/api/v1/authors/most-read"
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreAuthorStat
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MostPurchasedAuthor(ctx context.Context) (*model.BookStoreAuthorStat, error) {
	path := "/api/v1/authors/most-purchased"
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreAuthorStat
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Author(ctx context.Context, id int) (*model.BookStoreAuthor, error) {
	path := "/api/v1/authors/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreAuthor
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Publishers(ctx context.Context) ([]*model.BookStorePublisher, error) {
	path := "/api/v1/publishers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookStorePublisher
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Publisher(ctx context.Context, id int) (*model.BookStorePublisher, error) {
	path := "/api/v1/publishers/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStorePublisher
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context) ([]*model.BookStoreCategory, error) {
	path := "/api/v1/categories"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookStoreCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CategoryPriceStats(ctx context.Context) ([]*model.BookStoreCategoryPriceStat, error) {
	path := "/api/v1/categories/price-stats"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BookStoreCategoryPriceStat
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Category(ctx context.Context, id int) (*model.BookStoreCategory, error) {
	path := "/api/v1/categories/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookStoreCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
