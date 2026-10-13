package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) NotifCreateNotification(ctx context.Context, input model.NotifNotificationInput) (*model.NotifNotification, error) {
	return restCall[*model.NotifNotification](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/notifications", Body: input}, "")
}

func (r *mutationResolver) NotifUpdateNotification(ctx context.Context, id int, input model.NotifNotificationInput) (*model.NotifNotification, error) {
	return restCall[*model.NotifNotification](ctx, r.Proxy, "notification-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/notifications/%v", id), Body: input}, "")
}

func (r *mutationResolver) NotifDeleteNotification(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notification-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/notifications/%v", id)}, "")
}

func (r *mutationResolver) NotifSendEmail(ctx context.Context, input model.NotifEmailInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/notifications/email", Body: input}, "")
}

func (r *mutationResolver) NotifSendSlack(ctx context.Context, input model.NotifSlackInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/notifications/slack", Body: input}, "")
}

func (r *mutationResolver) NotifSendPush(ctx context.Context, input model.NotifPushInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/notifications/push", Body: input}, "")
}

func (r *mutationResolver) NotifCreateCampaign(ctx context.Context, input model.NotifCampaignInput) (*model.NotifCampaign, error) {
	return restCall[*model.NotifCampaign](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/campaigns", Body: input}, "")
}

func (r *mutationResolver) NotifCancelCampaign(ctx context.Context, id int) (*model.NotifCampaignView, error) {
	return restCall[*model.NotifCampaignView](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/campaigns/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) NotifUpsertContact(ctx context.Context, id int, email *string, phone *string) (*model.NotifUserContact, error) {
	return restCall[*model.NotifUserContact](ctx, r.Proxy, "notification-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/users/%v/contact", id), Body: map[string]any{"email": email, "phone": phone}}, "")
}

func (r *mutationResolver) NotifUpdateDevice(ctx context.Context, id int, input model.NotifDeviceUpdateInput) (*model.NotifDevice, error) {
	return restCall[*model.NotifDevice](ctx, r.Proxy, "notification-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/devices/%v", id), Body: input}, "")
}

func (r *mutationResolver) NotifDeleteDevice(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notification-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/devices/%v", id)}, "")
}

func (r *mutationResolver) RegisterMyDevice(ctx context.Context, input model.NotifDeviceInput) (*model.NotifDevice, error) {
	return restCall[*model.NotifDevice](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/devices", Body: withField(input, "user_id", userIDInt(ctx))}, "")
}

func (r *mutationResolver) UnregisterMyDevice(ctx context.Context, input model.NotifUnregisterDeviceInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "notification-service", rest.Request{Method: "POST", Path: "/api/v1/devices/unregister", Body: withField(input, "user_id", userIDInt(ctx))}, "")
}

func (r *mutationResolver) UpdateMyContact(ctx context.Context, email *string, phone *string) (*model.NotifUserContact, error) {
	return restCall[*model.NotifUserContact](ctx, r.Proxy, "notification-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/users/%v/contact", userIDOf(ctx)), Body: map[string]any{"email": email, "phone": phone}}, "")
}

func (r *queryResolver) Notifications(ctx context.Context) ([]*model.NotifNotification, error) {
	return restCall[[]*model.NotifNotification](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: "/api/v1/notifications"}, "")
}

func (r *queryResolver) Notification(ctx context.Context, id int) (*model.NotifNotification, error) {
	return restCall[*model.NotifNotification](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/notifications/%v", id)}, "")
}

func (r *queryResolver) NotifCampaigns(ctx context.Context, limit *int, offset *int) ([]*model.NotifCampaign, error) {
	return restCall[[]*model.NotifCampaign](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: "/api/v1/campaigns", Query: restQuery("limit", limit, "offset", offset)}, "items")
}

func (r *queryResolver) NotifCampaign(ctx context.Context, id int) (*model.NotifCampaignView, error) {
	return restCall[*model.NotifCampaignView](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/campaigns/%v", id)}, "")
}

func (r *queryResolver) NotifAuditDeliveries(ctx context.Context, email *string, phone *string, channel *string, status *string, sourceType *string, requestedBy *string, userID *int, sourceID *int, beforeID *int, limit *int, from *string, to *string) (*model.NotifAuditPage, error) {
	return restCall[*model.NotifAuditPage](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: "/api/v1/audit/deliveries", Query: restQuery("email", email, "phone", phone, "channel", channel, "status", status, "source_type", sourceType, "requested_by", requestedBy, "user_id", userID, "source_id", sourceID, "before_id", beforeID, "limit", limit, "from", from, "to", to)}, "")
}

func (r *queryResolver) NotifAuditContent(ctx context.Context, id int) (*model.NotifAuditContent, error) {
	return restCall[*model.NotifAuditContent](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/audit/contents/%v", id)}, "")
}

func (r *queryResolver) NotifUserContact(ctx context.Context, id int) (*model.NotifUserContact, error) {
	return restCall[*model.NotifUserContact](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/contact", id)}, "")
}

func (r *queryResolver) NotifDevices(ctx context.Context, userID int) ([]*model.NotifDevice, error) {
	return restCall[[]*model.NotifDevice](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: "/api/v1/devices", Query: restQuery("user_id", userID)}, "")
}

func (r *queryResolver) NotifDevice(ctx context.Context, id int) (*model.NotifDevice, error) {
	return restCall[*model.NotifDevice](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/devices/%v", id)}, "")
}

func (r *queryResolver) MyDevices(ctx context.Context) ([]*model.NotifDevice, error) {
	return restCall[[]*model.NotifDevice](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: "/api/v1/devices", Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *queryResolver) MyContact(ctx context.Context) (*model.NotifUserContact, error) {
	return restCall[*model.NotifUserContact](ctx, r.Proxy, "notification-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/contact", userIDOf(ctx))}, "")
}
