package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RoomQuery() generated.RoomQueryResolver { return &roomQueryResolver{r} }

type roomQueryResolver struct{ *Resolver }

func (r *roomQueryResolver) Me(ctx context.Context, obj *model.RoomQuery) (*model.RoomUser, error) {
	return r.Clients.RoomService.Me(ctx)
}

func (r *roomQueryResolver) Rooms(ctx context.Context, obj *model.RoomQuery) ([]*model.RoomRoom, error) {
	return r.Clients.RoomService.Rooms(ctx)
}

func (r *roomQueryResolver) Room(ctx context.Context, obj *model.RoomQuery, id int) (*model.RoomRoom, error) {
	return r.Clients.RoomService.Room(ctx, id)
}

func (r *roomQueryResolver) Members(ctx context.Context, obj *model.RoomQuery, id int) ([]*model.RoomMember, error) {
	return r.Clients.RoomService.Members(ctx, id)
}

func (r *roomQueryResolver) Messages(ctx context.Context, obj *model.RoomQuery, id int, before *int, limit *int) ([]*model.RoomMessage, error) {
	return r.Clients.RoomService.Messages(ctx, id, before, limit)
}
