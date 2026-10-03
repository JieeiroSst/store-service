package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) HubQuery() generated.HubQueryResolver { return &hubQueryResolver{r} }

type hubQueryResolver struct{ *Resolver }

func (r *hubQueryResolver) Channels(ctx context.Context, obj *model.HubQuery, typeArg *string, active *bool) ([]*model.HubChannel, error) {
	return r.Clients.NotifyhubService.Channels(ctx, typeArg, active)
}

func (r *hubQueryResolver) Channel(ctx context.Context, obj *model.HubQuery, id string) (*model.HubChannel, error) {
	return r.Clients.NotifyhubService.Channel(ctx, id)
}

func (r *hubQueryResolver) DataSources(ctx context.Context, obj *model.HubQuery) ([]*model.HubDataSource, error) {
	return r.Clients.NotifyhubService.DataSources(ctx)
}

func (r *hubQueryResolver) DataSource(ctx context.Context, obj *model.HubQuery, id string) (*model.HubDataSource, error) {
	return r.Clients.NotifyhubService.DataSource(ctx, id)
}

func (r *hubQueryResolver) Templates(ctx context.Context, obj *model.HubQuery, channel *string) ([]*model.HubTemplate, error) {
	return r.Clients.NotifyhubService.Templates(ctx, channel)
}

func (r *hubQueryResolver) Template(ctx context.Context, obj *model.HubQuery, id string) (*model.HubTemplate, error) {
	return r.Clients.NotifyhubService.Template(ctx, id)
}

func (r *hubQueryResolver) Jobs(ctx context.Context, obj *model.HubQuery, status *string, page *int, pageSize *int) (*model.HubJobPage, error) {
	return r.Clients.NotifyhubService.Jobs(ctx, status, page, pageSize)
}

func (r *hubQueryResolver) Job(ctx context.Context, obj *model.HubQuery, id string) (*model.HubNotifyJob, error) {
	return r.Clients.NotifyhubService.Job(ctx, id)
}

func (r *hubQueryResolver) History(ctx context.Context, obj *model.HubQuery, jobID *string, status *string, page *int, pageSize *int) (*model.HubHistoryPage, error) {
	return r.Clients.NotifyhubService.History(ctx, jobID, status, page, pageSize)
}

func (r *hubQueryResolver) SchedulerStatus(ctx context.Context, obj *model.HubQuery) (*model.HubSchedulerStatus, error) {
	return r.Clients.NotifyhubService.SchedulerStatus(ctx)
}
