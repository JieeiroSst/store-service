package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RecruitCreateJob(ctx context.Context, input model.RecruitJobInput) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: "/api/v1/jobs", Body: input}, "")
}

func (r *mutationResolver) RecruitUpdateJob(ctx context.Context, id string, input model.RecruitJobInput) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/jobs/%v", id), Body: input}, "")
}

func (r *mutationResolver) RecruitPublishJob(ctx context.Context, id string, reason *string) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/jobs/%v/publish", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) RecruitPauseJob(ctx context.Context, id string, reason *string) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/jobs/%v/pause", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) RecruitCloseJob(ctx context.Context, id string, reason *string) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/jobs/%v/close", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) RecruitApply(ctx context.Context, jobID string, candidateID string, referredByPartnerID *string, coverLetter *string) (*model.RecruitApplication, error) {
	return restCall[*model.RecruitApplication](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: "/api/v1/applications", Body: map[string]any{"job_id": jobID, "candidate_id": candidateID, "referred_by_partner_id": referredByPartnerID, "cover_letter": coverLetter}}, "")
}

func (r *mutationResolver) RecruitMoveStage(ctx context.Context, id string, stageID string, status string) (*model.RecruitApplication, error) {
	return restCall[*model.RecruitApplication](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/applications/%v/stage", id), Body: map[string]any{"stage_id": stageID, "status": status}}, "")
}

func (r *mutationResolver) RecruitScheduleInterview(ctx context.Context, id string, input model.RecruitInterviewInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/applications/%v/interviews", id), Body: input}, "")
}

func (r *mutationResolver) RecruitSubmitFeedback(ctx context.Context, id string, interviewID string, input model.RecruitFeedbackInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/applications/%v/interviews/%v/feedback", id, interviewID), Body: input}, "")
}

func (r *mutationResolver) RecruitExtendOffer(ctx context.Context, id string, input model.RecruitOfferInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/applications/%v/offer", id), Body: input}, "")
}

func (r *mutationResolver) RecruitReject(ctx context.Context, id string, reason string, note *string) (*model.RecruitApplication, error) {
	return restCall[*model.RecruitApplication](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/applications/%v/reject", id), Body: map[string]any{"reason": reason, "note": note}}, "")
}

func (r *mutationResolver) RecruitRegisterPartner(ctx context.Context, input model.RecruitPartnerInput) (*model.RecruitPartner, error) {
	return restCall[*model.RecruitPartner](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: "/api/v1/referrals/partners", Body: input}, "")
}

func (r *mutationResolver) RecruitRequestPayout(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/referrals/partners/%v/payout", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RecruitGenerateLink(ctx context.Context, partnerID string, jobID *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: "/api/v1/referrals/links", Body: map[string]any{"partner_id": partnerID, "job_id": jobID}}, "")
}

func (r *mutationResolver) RecruitCreateCandidate(ctx context.Context, input model.RecruitCandidateInput) (*model.RecruitCandidate, error) {
	return restCall[*model.RecruitCandidate](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: "/api/v1/candidates", Body: input}, "")
}

func (r *mutationResolver) RecruitUpdateCandidate(ctx context.Context, id string, input map[string]any) (*model.RecruitCandidate, error) {
	return restCall[*model.RecruitCandidate](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/candidates/%v", id), Body: input}, "")
}

func (r *mutationResolver) RecruitTransitionCandidate(ctx context.Context, id string, status string, reason *string) (*model.RecruitCandidate, error) {
	return restCall[*model.RecruitCandidate](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/candidates/%v/status", id), Body: map[string]any{"status": status, "reason": reason}}, "")
}

func (r *mutationResolver) RecruitEnrichCandidate(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/candidates/%v/enrich", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) RecruitJobs(ctx context.Context, page *int, limit *int, status *string, departmentID *string, recruiterID *int, workMode *string, skills []string, search *string) (*model.RecruitJobPage, error) {
	return restCall[*model.RecruitJobPage](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/jobs", Query: restQuery("page", page, "limit", limit, "status", status, "department_id", departmentID, "recruiter_id", recruiterID, "work_mode", workMode, "skills", skills, "search", search)}, "")
}

func (r *queryResolver) RecruitJob(ctx context.Context, id string) (*model.RecruitJob, error) {
	return restCall[*model.RecruitJob](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/jobs/%v", id)}, "")
}

func (r *queryResolver) RecruitRecommendedCandidates(ctx context.Context, id string) ([]*model.RecruitCandidate, error) {
	return restCall[[]*model.RecruitCandidate](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/jobs/%v/recommended-candidates", id)}, "candidates")
}

func (r *queryResolver) RecruitApplications(ctx context.Context, page *int, limit *int, jobID *string, candidateID *string, recruiterID *int, status *string) (*model.RecruitApplicationPage, error) {
	return restCall[*model.RecruitApplicationPage](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/applications", Query: restQuery("page", page, "limit", limit, "job_id", jobID, "candidate_id", candidateID, "recruiter_id", recruiterID, "status", status)}, "")
}

func (r *queryResolver) RecruitApplication(ctx context.Context, id string) (*model.RecruitApplication, error) {
	return restCall[*model.RecruitApplication](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/applications/%v", id)}, "")
}

func (r *queryResolver) RecruitCandidates(ctx context.Context, page *int, limit *int, status *string, source *string, experienceLevel *string, skills []string, minScore *float64, search *string) (*model.RecruitCandidatePage, error) {
	return restCall[*model.RecruitCandidatePage](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/candidates", Query: restQuery("page", page, "limit", limit, "status", status, "source", source, "experience_level", experienceLevel, "skills", skills, "min_score", minScore, "search", search)}, "")
}

func (r *queryResolver) RecruitCandidate(ctx context.Context, id string) (*model.RecruitCandidate, error) {
	return restCall[*model.RecruitCandidate](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/candidates/%v", id)}, "")
}

func (r *queryResolver) RecruitPartnerStats(ctx context.Context, id string) (*model.RecruitPartnerStats, error) {
	return restCall[*model.RecruitPartnerStats](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referrals/partners/%v/stats", id)}, "")
}

func (r *queryResolver) RecruitPartnerNetwork(ctx context.Context, id string) (*model.RecruitPartnerNetwork, error) {
	return restCall[*model.RecruitPartnerNetwork](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referrals/partners/%v/network", id)}, "")
}

func (r *queryResolver) RecruitReferralLeaderboard(ctx context.Context) ([]*model.RecruitPartnerStats, error) {
	return restCall[[]*model.RecruitPartnerStats](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/referrals/leaderboard"}, "leaderboard")
}

func (r *queryResolver) RecruitFunnel(ctx context.Context) (*model.RecruitFunnel, error) {
	return restCall[*model.RecruitFunnel](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/analytics/funnel"}, "")
}

func (r *queryResolver) RecruitFunnelByJob(ctx context.Context, jobID string) (*model.RecruitJobFunnel, error) {
	return restCall[*model.RecruitJobFunnel](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/analytics/funnel/%v", jobID)}, "")
}

func (r *queryResolver) RecruiterPerformance(ctx context.Context) ([]*model.RecruitRecruiterRow, error) {
	return restCall[[]*model.RecruitRecruiterRow](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/analytics/recruiter"}, "recruiters")
}

func (r *queryResolver) RecruitSLABreaches(ctx context.Context, breachedOnly *bool) (*model.RecruitSLABreaches, error) {
	return restCall[*model.RecruitSLABreaches](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/analytics/sla-breaches", Query: restQuery("breached_only", breachedOnly)}, "")
}

func (r *queryResolver) RecruitAiScoreDistribution(ctx context.Context) ([]*model.RecruitScoreRow, error) {
	return restCall[[]*model.RecruitScoreRow](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/analytics/ai-scores"}, "score_distribution")
}

func (r *queryResolver) RecruitReferralNetworkStats(ctx context.Context) ([]*model.RecruitNetworkRow, error) {
	return restCall[[]*model.RecruitNetworkRow](ctx, r.Proxy, "recruitment-platform-service", rest.Request{Method: "GET", Path: "/api/v1/analytics/referral-network"}, "network")
}
