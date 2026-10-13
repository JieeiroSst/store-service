package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ToolGenerateTestCases(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/testcases/generate", Body: input}, "")
}

func (r *mutationResolver) ToolRunTestCases(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/testcases/run", Body: input}, "")
}

func (r *mutationResolver) ToolRunBenchmark(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/benchmark", Body: input}, "")
}

func (r *mutationResolver) ToolAnalyzeCode(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/analyze/code", Body: input}, "")
}

func (r *mutationResolver) ToolAnalyzeTask(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/tasks/analyze", Body: input}, "")
}

func (r *mutationResolver) ToolQaScan(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/scan", Body: input}, "")
}

func (r *mutationResolver) ToolExportXlsx(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/export/xlsx", Body: input}, "")
}

func (r *mutationResolver) ToolQaDatabase(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/database", Body: input}, "")
}

func (r *mutationResolver) ToolQaAsync(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/async", Body: input}, "")
}

func (r *mutationResolver) ToolQaQueue(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/queue", Body: input}, "")
}

func (r *mutationResolver) ToolQaBrowser(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/browser", Body: input}, "")
}

func (r *mutationResolver) ToolQaActive(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/active", Body: input}, "")
}

func (r *mutationResolver) ToolJobStress(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/jobs/stress", Body: input}, "")
}

func (r *mutationResolver) ToolJobSoak(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/jobs/soak", Body: input}, "")
}

func (r *mutationResolver) ToolMobileStatic(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/mobile/static", Body: input}, "")
}

func (r *mutationResolver) ToolMobileAndroid(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/mobile/android", Body: input}, "")
}

func (r *mutationResolver) ToolMobileIos(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qa/mobile/ios", Body: input}, "")
}

func (r *mutationResolver) ToolQcRun(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/qc/run", Body: input}, "")
}

func (r *mutationResolver) ToolLearningFeedback(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/learning/feedback", Body: input}, "")
}

func (r *mutationResolver) ToolLearningReflect(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/learning/reflect", Body: input}, "")
}

func (r *mutationResolver) ToolLearningEvolve(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: "/api/v1/learning/evolve", Body: input}, "")
}

func (r *mutationResolver) ToolCancelJob(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/jobs/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ToolApproveCases(ctx context.Context, name string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/suites/%v/approve", name), Body: input}, "")
}

func (r *mutationResolver) ToolRemoveCase(ctx context.Context, name string, hash string) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/suites/%v/cases/%v", name, hash)}, "")
}

func (r *mutationResolver) ToolRunSuite(ctx context.Context, name string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "tool-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/suites/%v/run", name), Body: input}, "")
}

func (r *queryResolver) ToolJobs(ctx context.Context) ([]*model.ToolJob, error) {
	return restCall[[]*model.ToolJob](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: "/api/v1/jobs"}, "jobs")
}

func (r *queryResolver) ToolJob(ctx context.Context, id string) (*model.ToolJob, error) {
	return restCall[*model.ToolJob](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/jobs/%v", id)}, "")
}

func (r *queryResolver) ToolMobileDevices(ctx context.Context) (*model.ToolMobileDevices, error) {
	return restCall[*model.ToolMobileDevices](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: "/api/v1/qa/mobile/devices"}, "")
}

func (r *queryResolver) ToolLearningStats(ctx context.Context) (*model.ToolStats, error) {
	return restCall[*model.ToolStats](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: "/api/v1/learning/stats"}, "")
}

func (r *queryResolver) ToolLearningPending(ctx context.Context) ([]*model.ToolExperience, error) {
	return restCall[[]*model.ToolExperience](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: "/api/v1/learning/pending"}, "items")
}

func (r *queryResolver) ToolLearningHistory(ctx context.Context) ([]*model.ToolSnapshot, error) {
	return restCall[[]*model.ToolSnapshot](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: "/api/v1/learning/history"}, "history")
}

func (r *queryResolver) ToolSuite(ctx context.Context, name string) (*model.ToolSuiteResult, error) {
	return restCall[*model.ToolSuiteResult](ctx, r.Proxy, "tool-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/suites/%v", name)}, "")
}
