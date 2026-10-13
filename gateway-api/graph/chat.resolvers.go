package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CreateChatReport(ctx context.Context, input model.ChatReportInput) (*model.ChatCreatedReport, error) {
	return restCall[*model.ChatCreatedReport](ctx, r.Proxy, "chat-service", rest.Request{Method: "POST", Path: "/report", Body: withField(input, "UserId", userIDInt(ctx))}, "")
}

func (r *mutationResolver) DeleteMyChatMessage(ctx context.Context, messageID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "chat-service", rest.Request{Method: "DELETE", Path: rest.P("/message/%v/%v", userIDOf(ctx), messageID)}, "")
}

func (r *queryResolver) ChatMessage(ctx context.Context, id int) (*model.ChatMessage, error) {
	return restCall[*model.ChatMessage](ctx, r.Proxy, "chat-service", rest.Request{Method: "GET", Path: rest.P("/message/%v", id)}, "")
}

func (r *queryResolver) MyChatReports(ctx context.Context) ([]*model.ChatReport, error) {
	return restCall[[]*model.ChatReport](ctx, r.Proxy, "chat-service", rest.Request{Method: "GET", Path: rest.P("/report/%v", userIDOf(ctx))}, "")
}

func (r *queryResolver) ChatUserReports(ctx context.Context, userID int) ([]*model.ChatReport, error) {
	return restCall[[]*model.ChatReport](ctx, r.Proxy, "chat-service", rest.Request{Method: "GET", Path: rest.P("/report/%v", userID)}, "")
}
