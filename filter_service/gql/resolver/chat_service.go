package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ChatQuery() generated.ChatQueryResolver { return &chatQueryResolver{r} }

type chatQueryResolver struct{ *Resolver }

func (r *chatQueryResolver) Message(ctx context.Context, obj *model.ChatQuery, id int) (*model.ChatMessage, error) {
	return r.Clients.ChatService.Message(ctx, id)
}

func (r *chatQueryResolver) UserReports(ctx context.Context, obj *model.ChatQuery, userID int) ([]*model.ChatReport, error) {
	return r.Clients.ChatService.UserReports(ctx, userID)
}
