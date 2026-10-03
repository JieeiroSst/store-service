package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) OllamaChatQuery() generated.OllamaChatQueryResolver {
	return &ollamaChatQueryResolver{r}
}

type ollamaChatQueryResolver struct{ *Resolver }

func (r *ollamaChatQueryResolver) Contacts(ctx context.Context, obj *model.OllamaChatQuery) (*model.OllamaChatContacts, error) {
	return r.Clients.OllamaService.Contacts(ctx)
}

func (r *ollamaChatQueryResolver) SearchGroups(ctx context.Context, obj *model.OllamaChatQuery, qArg string) (*model.OllamaChatGroupSearch, error) {
	return r.Clients.OllamaService.SearchGroups(ctx, qArg)
}

func (r *ollamaChatQueryResolver) Groups(ctx context.Context, obj *model.OllamaChatQuery) (*model.OllamaChatGroups, error) {
	return r.Clients.OllamaService.Groups(ctx)
}

func (r *ollamaChatQueryResolver) GroupMembers(ctx context.Context, obj *model.OllamaChatQuery, groupID int) (*model.OllamaChatGroupMembers, error) {
	return r.Clients.OllamaService.GroupMembers(ctx, groupID)
}

func (r *ollamaChatQueryResolver) ChatHistory(ctx context.Context, obj *model.OllamaChatQuery, limit *int, offset *int, groupID *int, recipientID *int) (*model.OllamaChatHistory, error) {
	return r.Clients.OllamaService.ChatHistory(ctx, limit, offset, groupID, recipientID)
}
