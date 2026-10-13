package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *queryResolver) WebrtcIceServers(ctx context.Context, userID *string) (*model.WebrtcIceServers, error) {
	return restCall[*model.WebrtcIceServers](ctx, r.Proxy, "webrtc-service", rest.Request{Method: "GET", Path: "/api/ice-servers", Query: restQuery("user_id", userID)}, "")
}

func (r *queryResolver) WebrtcRoom(ctx context.Context, roomID string) (*model.WebrtcRoomInfo, error) {
	return restCall[*model.WebrtcRoomInfo](ctx, r.Proxy, "webrtc-service", rest.Request{Method: "GET", Path: rest.P("/api/rooms/%v", roomID)}, "")
}
