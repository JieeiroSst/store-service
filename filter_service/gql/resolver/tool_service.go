package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ToolQuery() generated.ToolQueryResolver { return &toolQueryResolver{r} }

type toolQueryResolver struct{ *Resolver }

func (r *toolQueryResolver) Jobs(ctx context.Context, obj *model.ToolQuery) ([]*model.ToolJob, error) {
	return r.Clients.ToolService.Jobs(ctx)
}

func (r *toolQueryResolver) Job(ctx context.Context, obj *model.ToolQuery, id string) (*model.ToolJob, error) {
	return r.Clients.ToolService.Job(ctx, id)
}

func (r *toolQueryResolver) MobileDevices(ctx context.Context, obj *model.ToolQuery) (*model.ToolMobileDevices, error) {
	return r.Clients.ToolService.MobileDevices(ctx)
}

func (r *toolQueryResolver) LearningStats(ctx context.Context, obj *model.ToolQuery) (*model.ToolStats, error) {
	return r.Clients.ToolService.LearningStats(ctx)
}

func (r *toolQueryResolver) LearningPending(ctx context.Context, obj *model.ToolQuery) ([]*model.ToolExperience, error) {
	return r.Clients.ToolService.LearningPending(ctx)
}

func (r *toolQueryResolver) LearningHistory(ctx context.Context, obj *model.ToolQuery) ([]*model.ToolSnapshot, error) {
	return r.Clients.ToolService.LearningHistory(ctx)
}

func (r *toolQueryResolver) Suite(ctx context.Context, obj *model.ToolQuery, name string) (*model.ToolSuiteResult, error) {
	return r.Clients.ToolService.Suite(ctx, name)
}
