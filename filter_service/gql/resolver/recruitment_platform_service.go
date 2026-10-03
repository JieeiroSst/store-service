package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RecruitQuery() generated.RecruitQueryResolver { return &recruitQueryResolver{r} }

type recruitQueryResolver struct{ *Resolver }

func (r *recruitQueryResolver) Jobs(ctx context.Context, obj *model.RecruitQuery, page *int, limit *int, status *string, departmentID *string, recruiterID *int, workMode *string, skills []string, search *string) (*model.RecruitJobPage, error) {
	return r.Clients.RecruitmentPlatformService.Jobs(ctx, page, limit, status, departmentID, recruiterID, workMode, skills, search)
}

func (r *recruitQueryResolver) Job(ctx context.Context, obj *model.RecruitQuery, id string) (*model.RecruitJob, error) {
	return r.Clients.RecruitmentPlatformService.Job(ctx, id)
}

func (r *recruitQueryResolver) RecommendedCandidates(ctx context.Context, obj *model.RecruitQuery, id string) ([]*model.RecruitCandidate, error) {
	return r.Clients.RecruitmentPlatformService.RecommendedCandidates(ctx, id)
}

func (r *recruitQueryResolver) Applications(ctx context.Context, obj *model.RecruitQuery, page *int, limit *int, jobID *string, candidateID *string, recruiterID *int, status *string) (*model.RecruitApplicationPage, error) {
	return r.Clients.RecruitmentPlatformService.Applications(ctx, page, limit, jobID, candidateID, recruiterID, status)
}

func (r *recruitQueryResolver) Application(ctx context.Context, obj *model.RecruitQuery, id string) (*model.RecruitApplication, error) {
	return r.Clients.RecruitmentPlatformService.Application(ctx, id)
}

func (r *recruitQueryResolver) Candidates(ctx context.Context, obj *model.RecruitQuery, page *int, limit *int, status *string, source *string, experienceLevel *string, skills []string, minScore *float64, search *string) (*model.RecruitCandidatePage, error) {
	return r.Clients.RecruitmentPlatformService.Candidates(ctx, page, limit, status, source, experienceLevel, skills, minScore, search)
}

func (r *recruitQueryResolver) Candidate(ctx context.Context, obj *model.RecruitQuery, id string) (*model.RecruitCandidate, error) {
	return r.Clients.RecruitmentPlatformService.Candidate(ctx, id)
}

func (r *recruitQueryResolver) PartnerStats(ctx context.Context, obj *model.RecruitQuery, id string) (*model.RecruitPartnerStats, error) {
	return r.Clients.RecruitmentPlatformService.PartnerStats(ctx, id)
}

func (r *recruitQueryResolver) PartnerNetwork(ctx context.Context, obj *model.RecruitQuery, id string) (*model.RecruitPartnerNetwork, error) {
	return r.Clients.RecruitmentPlatformService.PartnerNetwork(ctx, id)
}

func (r *recruitQueryResolver) ReferralLeaderboard(ctx context.Context, obj *model.RecruitQuery) ([]*model.RecruitPartnerStats, error) {
	return r.Clients.RecruitmentPlatformService.ReferralLeaderboard(ctx)
}

func (r *recruitQueryResolver) Funnel(ctx context.Context, obj *model.RecruitQuery) (*model.RecruitFunnel, error) {
	return r.Clients.RecruitmentPlatformService.Funnel(ctx)
}

func (r *recruitQueryResolver) FunnelByJob(ctx context.Context, obj *model.RecruitQuery, jobID string) (*model.RecruitJobFunnel, error) {
	return r.Clients.RecruitmentPlatformService.FunnelByJob(ctx, jobID)
}

func (r *recruitQueryResolver) RecruiterPerformance(ctx context.Context, obj *model.RecruitQuery) ([]*model.RecruitRecruiterRow, error) {
	return r.Clients.RecruitmentPlatformService.RecruiterPerformance(ctx)
}

func (r *recruitQueryResolver) SLABreaches(ctx context.Context, obj *model.RecruitQuery, breachedOnly *bool) (*model.RecruitSLABreaches, error) {
	return r.Clients.RecruitmentPlatformService.SLABreaches(ctx, breachedOnly)
}

func (r *recruitQueryResolver) AiScoreDistribution(ctx context.Context, obj *model.RecruitQuery) ([]*model.RecruitScoreRow, error) {
	return r.Clients.RecruitmentPlatformService.AiScoreDistribution(ctx)
}

func (r *recruitQueryResolver) ReferralNetworkStats(ctx context.Context, obj *model.RecruitQuery) ([]*model.RecruitNetworkRow, error) {
	return r.Clients.RecruitmentPlatformService.ReferralNetworkStats(ctx)
}
