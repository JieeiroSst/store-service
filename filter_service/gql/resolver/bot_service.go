package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BotQuery() generated.BotQueryResolver { return &botQueryResolver{r} }

type botQueryResolver struct{ *Resolver }

func (r *botQueryResolver) Posts(ctx context.Context, obj *model.BotQuery, limit *int, offset *int, status *string, campaign *string) ([]*model.BotPost, error) {
	return r.Clients.BotService.Posts(ctx, limit, offset, status, campaign)
}

func (r *botQueryResolver) Post(ctx context.Context, obj *model.BotQuery, id string) (*model.BotPost, error) {
	return r.Clients.BotService.Post(ctx, id)
}
