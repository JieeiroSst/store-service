package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) LiveCreateRoom(ctx context.Context, input model.LiveCreateRoomInput) (*model.LiveRoom, error) {
	return restCall[*model.LiveRoom](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: "/api/v1/rooms", Body: input}, "")
}

func (r *mutationResolver) LiveRegenerateStreamKey(ctx context.Context, id string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/stream-key/regenerate", id), Body: map[string]any{}}, "streamKey")
}

func (r *mutationResolver) LiveRequestIngest(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/ingest", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) LiveEndStream(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/end", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) LiveBanChat(ctx context.Context, id string, targetUserID string, durationSeconds *int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/chat/ban", id), Body: map[string]any{"targetUserId": targetUserID, "durationSeconds": durationSeconds}}, "")
}

func (r *mutationResolver) LiveUnbanChat(ctx context.Context, id string, targetUserID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/chat/unban", id), Body: map[string]any{"targetUserId": targetUserID}}, "")
}

func (r *mutationResolver) LiveViewerHeartbeat(ctx context.Context, id string, sessionID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/viewers/heartbeat", id), Body: map[string]any{"sessionId": sessionID}}, "")
}

func (r *mutationResolver) LiveReportQoE(ctx context.Context, id string, bitrateKbps *float64, bufferingEvents *int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/rooms/%v/qoe", id), Body: map[string]any{"bitrateKbps": bitrateKbps, "bufferingEvents": bufferingEvents}}, "")
}

func (r *mutationResolver) LiveAdminDeleteRoom(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "livestream-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/rooms/%v", id)}, "")
}

func (r *queryResolver) LiveRooms(ctx context.Context, live *bool) ([]*model.LiveRoom, error) {
	return restCall[[]*model.LiveRoom](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: "/api/v1/rooms", Query: restQuery("live", live)}, "")
}

func (r *queryResolver) LiveRoom(ctx context.Context, id string) (*model.LiveRoom, error) {
	return restCall[*model.LiveRoom](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/rooms/%v", id)}, "")
}

func (r *queryResolver) LiveActiveStream(ctx context.Context, id string) (*model.LiveStream, error) {
	return restCall[*model.LiveStream](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/rooms/%v/stream", id)}, "")
}

func (r *queryResolver) LiveRecordings(ctx context.Context, id string) ([]*model.LiveRecording, error) {
	return restCall[[]*model.LiveRecording](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/rooms/%v/recordings", id)}, "")
}

func (r *queryResolver) LiveViewerCount(ctx context.Context, id string) (*model.LiveViewerCount, error) {
	return restCall[*model.LiveViewerCount](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/rooms/%v/viewers", id)}, "")
}

func (r *queryResolver) LivePlayback(ctx context.Context, id string) (*model.LivePlaybackInfo, error) {
	return restCall[*model.LivePlaybackInfo](ctx, r.Proxy, "livestream-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/rooms/%v/playback", id)}, "")
}
