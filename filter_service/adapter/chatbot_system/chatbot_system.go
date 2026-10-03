package chatbot_system

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "chatbot-system"

const DefaultBaseURL = "http://chatbot-system:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) ConversationHistory(ctx context.Context, conversationID int, limit *int, offset *int) (*model.ChatbotMessageList, error) {
	path := "/api/conversations/" + url.PathEscape(strconv.Itoa(conversationID)) + "/history"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.ChatbotMessageList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Conversations(ctx context.Context) (*model.ChatbotConversationList, error) {
	path := "/api/conversations"
	q := url.Values{}
	h := http.Header{}
	var out *model.ChatbotConversationList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
