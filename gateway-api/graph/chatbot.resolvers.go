package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *queryResolver) ChatbotConversationHistory(ctx context.Context, conversationID int, limit *int, offset *int) (*model.ChatbotHistory, error) {
	return restCall[*model.ChatbotHistory](ctx, r.Proxy, "chatbot-system", rest.Request{Method: "GET", Path: rest.P("/api/conversations/%v/history", conversationID), Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) ChatbotConversations(ctx context.Context) (*model.ChatbotConversationList, error) {
	return restCall[*model.ChatbotConversationList](ctx, r.Proxy, "chatbot-system", rest.Request{Method: "GET", Path: "/api/conversations"}, "")
}
