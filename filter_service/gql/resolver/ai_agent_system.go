package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AiAgentQuery() generated.AiAgentQueryResolver { return &aiAgentQueryResolver{r} }

type aiAgentQueryResolver struct{ *Resolver }

func (r *aiAgentQueryResolver) ChatHistory(ctx context.Context, obj *model.AiAgentQuery, limit *int, xUserID string) (*model.AiAgentHistory, error) {
	return r.Clients.AiAgentSystem.ChatHistory(ctx, limit, xUserID)
}

func (r *aiAgentQueryResolver) Tags(ctx context.Context, obj *model.AiAgentQuery) (*model.AiAgentTags, error) {
	return r.Clients.AiAgentSystem.Tags(ctx)
}
