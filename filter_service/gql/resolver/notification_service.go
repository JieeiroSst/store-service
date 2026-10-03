package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) NotifQuery() generated.NotifQueryResolver { return &notifQueryResolver{r} }

type notifQueryResolver struct{ *Resolver }

func (r *notifQueryResolver) Notifications(ctx context.Context, obj *model.NotifQuery) ([]*model.NotifNotification, error) {
	return r.Clients.NotificationService.Notifications(ctx)
}

func (r *notifQueryResolver) Notification(ctx context.Context, obj *model.NotifQuery, id int) (*model.NotifNotification, error) {
	return r.Clients.NotificationService.Notification(ctx, id)
}

func (r *notifQueryResolver) Campaigns(ctx context.Context, obj *model.NotifQuery, limit *int, offset *int) ([]*model.NotifCampaign, error) {
	return r.Clients.NotificationService.Campaigns(ctx, limit, offset)
}

func (r *notifQueryResolver) Campaign(ctx context.Context, obj *model.NotifQuery, id int) (*model.NotifCampaignView, error) {
	return r.Clients.NotificationService.Campaign(ctx, id)
}

func (r *notifQueryResolver) AuditDeliveries(ctx context.Context, obj *model.NotifQuery, email *string, phone *string, channel *string, status *string, sourceType *string, requestedBy *string, userID *int, sourceID *int, beforeID *int, limit *int, from *string, to *string) (*model.NotifAuditPage, error) {
	return r.Clients.NotificationService.AuditDeliveries(ctx, email, phone, channel, status, sourceType, requestedBy, userID, sourceID, beforeID, limit, from, to)
}

func (r *notifQueryResolver) AuditContent(ctx context.Context, obj *model.NotifQuery, id int) (*model.NotifAuditContent, error) {
	return r.Clients.NotificationService.AuditContent(ctx, id)
}

func (r *notifQueryResolver) UserContact(ctx context.Context, obj *model.NotifQuery, id int) (*model.NotifUserContact, error) {
	return r.Clients.NotificationService.UserContact(ctx, id)
}

func (r *notifQueryResolver) Devices(ctx context.Context, obj *model.NotifQuery, userID int) ([]*model.NotifDevice, error) {
	return r.Clients.NotificationService.Devices(ctx, userID)
}

func (r *notifQueryResolver) Device(ctx context.Context, obj *model.NotifQuery, id int) (*model.NotifDevice, error) {
	return r.Clients.NotificationService.Device(ctx, id)
}
