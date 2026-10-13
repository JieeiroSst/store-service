package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) HubCreateChannel(ctx context.Context, input model.HubChannelInput) (*model.HubChannel, error) {
	return restCall[*model.HubChannel](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: "/api/v1/channels", Body: input}, "data")
}

func (r *mutationResolver) HubUpdateChannel(ctx context.Context, id string, input model.HubChannelInput) (*model.HubChannel, error) {
	return restCall[*model.HubChannel](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/channels/%v", id), Body: input}, "data")
}

func (r *mutationResolver) HubDeleteChannel(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/channels/%v", id)}, "")
}

func (r *mutationResolver) HubCreateDataSource(ctx context.Context, input model.HubDataSourceInput) (*model.HubDataSource, error) {
	return restCall[*model.HubDataSource](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: "/api/v1/data-sources", Body: input}, "data")
}

func (r *mutationResolver) HubUpdateDataSource(ctx context.Context, id string, input model.HubDataSourceInput) (*model.HubDataSource, error) {
	return restCall[*model.HubDataSource](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/data-sources/%v", id), Body: input}, "data")
}

func (r *mutationResolver) HubDeleteDataSource(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/data-sources/%v", id)}, "")
}

func (r *mutationResolver) HubCreateTemplate(ctx context.Context, input model.HubTemplateInput) (*model.HubTemplate, error) {
	return restCall[*model.HubTemplate](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: "/api/v1/templates", Body: input}, "data")
}

func (r *mutationResolver) HubUpdateTemplate(ctx context.Context, id string, input model.HubTemplateInput) (*model.HubTemplate, error) {
	return restCall[*model.HubTemplate](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/templates/%v", id), Body: input}, "data")
}

func (r *mutationResolver) HubDeleteTemplate(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/templates/%v", id)}, "")
}

func (r *mutationResolver) HubCreateJob(ctx context.Context, input model.HubNotifyJobInput) (*model.HubNotifyJob, error) {
	return restCall[*model.HubNotifyJob](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: "/api/v1/jobs", Body: input}, "data")
}

func (r *mutationResolver) HubUpdateJob(ctx context.Context, id string, input model.HubNotifyJobInput) (*model.HubNotifyJob, error) {
	return restCall[*model.HubNotifyJob](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/jobs/%v", id), Body: input}, "data")
}

func (r *mutationResolver) HubDeleteJob(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/jobs/%v", id)}, "")
}

func (r *mutationResolver) HubPauseJob(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/jobs/%v/pause", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) HubResumeJob(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/jobs/%v/resume", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) HubTriggerJob(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/jobs/%v/trigger", id), Body: map[string]any{}}, "data")
}

func (r *queryResolver) HubChannels(ctx context.Context, typeArg *string, active *bool) ([]*model.HubChannel, error) {
	return restCall[[]*model.HubChannel](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/channels", Query: restQuery("type", typeArg, "active", active)}, "data")
}

func (r *queryResolver) HubChannel(ctx context.Context, id string) (*model.HubChannel, error) {
	return restCall[*model.HubChannel](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/channels/%v", id)}, "data")
}

func (r *queryResolver) HubDataSources(ctx context.Context) ([]*model.HubDataSource, error) {
	return restCall[[]*model.HubDataSource](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/data-sources"}, "data")
}

func (r *queryResolver) HubDataSource(ctx context.Context, id string) (*model.HubDataSource, error) {
	return restCall[*model.HubDataSource](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/data-sources/%v", id)}, "data")
}

func (r *queryResolver) HubTemplates(ctx context.Context, channel *string) ([]*model.HubTemplate, error) {
	return restCall[[]*model.HubTemplate](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/templates", Query: restQuery("channel", channel)}, "data")
}

func (r *queryResolver) HubTemplate(ctx context.Context, id string) (*model.HubTemplate, error) {
	return restCall[*model.HubTemplate](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/templates/%v", id)}, "data")
}

func (r *queryResolver) HubJobs(ctx context.Context, status *string, page *int, pageSize *int) (*model.HubJobPage, error) {
	return restCall[*model.HubJobPage](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/jobs", Query: restQuery("status", status, "page", page, "page_size", pageSize)}, "")
}

func (r *queryResolver) HubJob(ctx context.Context, id string) (*model.HubNotifyJob, error) {
	return restCall[*model.HubNotifyJob](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/jobs/%v", id)}, "data")
}

func (r *queryResolver) HubHistory(ctx context.Context, jobID *string, status *string, page *int, pageSize *int) (*model.HubHistoryPage, error) {
	return restCall[*model.HubHistoryPage](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/history", Query: restQuery("job_id", jobID, "status", status, "page", page, "page_size", pageSize)}, "")
}

func (r *queryResolver) HubSchedulerStatus(ctx context.Context) (*model.HubSchedulerStatus, error) {
	return restCall[*model.HubSchedulerStatus](ctx, r.Proxy, "notifyhub-service", rest.Request{Method: "GET", Path: "/api/v1/scheduler/status"}, "")
}
