package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RoomCreateRoom(ctx context.Context, name string) (*model.RoomRoom, error) {
	return restCall[*model.RoomRoom](ctx, r.Proxy, "room-service", rest.Request{Method: "POST", Path: "/api/rooms", Body: map[string]any{"name": name}}, "")
}

func (r *mutationResolver) RoomAddMember(ctx context.Context, id string, username string) (any, error) {
	return restCall[any](ctx, r.Proxy, "room-service", rest.Request{Method: "POST", Path: rest.P("/api/rooms/%v/members", id), Body: map[string]any{"username": username}}, "")
}

func (r *mutationResolver) RoomRemoveMember(ctx context.Context, id string, username string) (any, error) {
	return restCall[any](ctx, r.Proxy, "room-service", rest.Request{Method: "DELETE", Path: rest.P("/api/rooms/%v/members/%v", id, username)}, "")
}

func (r *queryResolver) RoomMe(ctx context.Context) (*model.RoomUser, error) {
	return restCall[*model.RoomUser](ctx, r.Proxy, "room-service", rest.Request{Method: "GET", Path: "/api/me"}, "")
}

func (r *queryResolver) Rooms(ctx context.Context) ([]*model.RoomRoom, error) {
	return restCall[[]*model.RoomRoom](ctx, r.Proxy, "room-service", rest.Request{Method: "GET", Path: "/api/rooms"}, "")
}

func (r *queryResolver) Room(ctx context.Context, id int) (*model.RoomRoom, error) {
	return restCall[*model.RoomRoom](ctx, r.Proxy, "room-service", rest.Request{Method: "GET", Path: rest.P("/api/rooms/%v", id)}, "")
}

func (r *queryResolver) RoomMembers(ctx context.Context, id int) ([]*model.RoomMember, error) {
	return restCall[[]*model.RoomMember](ctx, r.Proxy, "room-service", rest.Request{Method: "GET", Path: rest.P("/api/rooms/%v/members", id)}, "")
}

func (r *queryResolver) RoomMessages(ctx context.Context, id int, before *int, limit *int) ([]*model.RoomMessage, error) {
	return restCall[[]*model.RoomMessage](ctx, r.Proxy, "room-service", rest.Request{Method: "GET", Path: rest.P("/api/rooms/%v/messages", id), Query: restQuery("before", before, "limit", limit)}, "")
}
