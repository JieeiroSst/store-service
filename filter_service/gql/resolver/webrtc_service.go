package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) WebrtcQuery() generated.WebrtcQueryResolver { return &webrtcQueryResolver{r} }

type webrtcQueryResolver struct{ *Resolver }

func (r *webrtcQueryResolver) IceServers(ctx context.Context, obj *model.WebrtcQuery, userID *string) (*model.WebrtcIceServers, error) {
	return r.Clients.WebrtcService.IceServers(ctx, userID)
}

func (r *webrtcQueryResolver) Room(ctx context.Context, obj *model.WebrtcQuery, roomID string) (*model.WebrtcRoomInfo, error) {
	return r.Clients.WebrtcService.Room(ctx, roomID)
}
