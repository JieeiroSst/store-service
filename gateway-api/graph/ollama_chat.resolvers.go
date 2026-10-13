package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) OllamaChatCreateGroup(ctx context.Context, name string, description *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-service", rest.Request{Method: "POST", Path: "/groups/create", Body: map[string]any{"name": name, "description": description}}, "")
}

func (r *mutationResolver) OllamaChatInviteToGroup(ctx context.Context, groupID int, userID int, role string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-service", rest.Request{Method: "POST", Path: "/groups/invite", Body: map[string]any{"group_id": groupID, "user_id": userID, "role": role}}, "")
}

func (r *mutationResolver) OllamaChatLeaveGroup(ctx context.Context, groupID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-service", rest.Request{Method: "DELETE", Path: rest.P("/groups/%v/leave", groupID)}, "")
}

func (r *mutationResolver) OllamaChatPromoteMember(ctx context.Context, groupID int, userID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-service", rest.Request{Method: "POST", Path: rest.P("/groups/%v/members/%v/promote", groupID, userID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) OllamaChatRemoveMember(ctx context.Context, groupID int, userID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-service", rest.Request{Method: "DELETE", Path: rest.P("/groups/%v/members/%v", groupID, userID)}, "")
}

func (r *queryResolver) OllamaChatContacts(ctx context.Context) (*model.OllamaChatContacts, error) {
	return restCall[*model.OllamaChatContacts](ctx, r.Proxy, "ollama-service", rest.Request{Method: "GET", Path: "/users/contacts"}, "")
}

func (r *queryResolver) OllamaChatSearchGroups(ctx context.Context, q string) (*model.OllamaChatGroupSearch, error) {
	return restCall[*model.OllamaChatGroupSearch](ctx, r.Proxy, "ollama-service", rest.Request{Method: "GET", Path: "/users/search", Query: restQuery("q", q)}, "")
}

func (r *queryResolver) OllamaChatGroups(ctx context.Context) (*model.OllamaChatGroups, error) {
	return restCall[*model.OllamaChatGroups](ctx, r.Proxy, "ollama-service", rest.Request{Method: "GET", Path: "/groups/"}, "")
}

func (r *queryResolver) OllamaChatGroupMembers(ctx context.Context, groupID int) (*model.OllamaChatGroupMembers, error) {
	return restCall[*model.OllamaChatGroupMembers](ctx, r.Proxy, "ollama-service", rest.Request{Method: "GET", Path: rest.P("/groups/%v/members", groupID)}, "")
}

func (r *queryResolver) OllamaChatChatHistory(ctx context.Context, limit *int, offset *int, groupID *int, recipientID *int) (*model.OllamaChatHistory, error) {
	return restCall[*model.OllamaChatHistory](ctx, r.Proxy, "ollama-service", rest.Request{Method: "GET", Path: "/chat/history", Query: restQuery("limit", limit, "offset", offset, "group_id", groupID, "recipient_id", recipientID)}, "")
}
