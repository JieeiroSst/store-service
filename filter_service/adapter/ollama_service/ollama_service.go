package ollama_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "ollama-service"

const DefaultBaseURL = "http://ollama-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Contacts(ctx context.Context) (*model.OllamaChatContacts, error) {
	path := "/users/contacts"
	q := url.Values{}
	h := http.Header{}
	var out *model.OllamaChatContacts
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SearchGroups(ctx context.Context, qArg string) (*model.OllamaChatGroupSearch, error) {
	path := "/users/search"
	q := url.Values{}
	q.Set("q", qArg)
	h := http.Header{}
	var out *model.OllamaChatGroupSearch
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Groups(ctx context.Context) (*model.OllamaChatGroups, error) {
	path := "/groups/"
	q := url.Values{}
	h := http.Header{}
	var out *model.OllamaChatGroups
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupMembers(ctx context.Context, groupID int) (*model.OllamaChatGroupMembers, error) {
	path := "/groups/" + url.PathEscape(strconv.Itoa(groupID)) + "/members"
	q := url.Values{}
	h := http.Header{}
	var out *model.OllamaChatGroupMembers
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ChatHistory(ctx context.Context, limit *int, offset *int, groupID *int, recipientID *int) (*model.OllamaChatHistory, error) {
	path := "/chat/history"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if groupID != nil {
		q.Set("group_id", strconv.Itoa(*groupID))
	}
	if recipientID != nil {
		q.Set("recipient_id", strconv.Itoa(*recipientID))
	}
	h := http.Header{}
	var out *model.OllamaChatHistory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
