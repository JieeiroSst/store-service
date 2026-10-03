package recruitment_platform_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "recruitment-platform-service"

const DefaultBaseURL = "http://recruitment-platform-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Jobs(ctx context.Context, page *int, limit *int, status *string, departmentID *string, recruiterID *int, workMode *string, skills []string, search *string) (*model.RecruitJobPage, error) {
	path := "/api/v1/jobs"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if departmentID != nil {
		q.Set("department_id", *departmentID)
	}
	if recruiterID != nil {
		q.Set("recruiter_id", strconv.Itoa(*recruiterID))
	}
	if workMode != nil {
		q.Set("work_mode", *workMode)
	}
	for _, e := range skills {
		q.Add("skills", e)
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out *model.RecruitJobPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Job(ctx context.Context, id string) (*model.RecruitJob, error) {
	path := "/api/v1/jobs/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitJob
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RecommendedCandidates(ctx context.Context, id string) ([]*model.RecruitCandidate, error) {
	path := "/api/v1/jobs/" + url.PathEscape(id) + "/recommended-candidates"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RecruitCandidate
	err := c.rest.Get(ctx, path, q, h, "candidates", &out)
	return out, err
}

func (c *Client) Applications(ctx context.Context, page *int, limit *int, jobID *string, candidateID *string, recruiterID *int, status *string) (*model.RecruitApplicationPage, error) {
	path := "/api/v1/applications"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if jobID != nil {
		q.Set("job_id", *jobID)
	}
	if candidateID != nil {
		q.Set("candidate_id", *candidateID)
	}
	if recruiterID != nil {
		q.Set("recruiter_id", strconv.Itoa(*recruiterID))
	}
	if status != nil {
		q.Set("status", *status)
	}
	h := http.Header{}
	var out *model.RecruitApplicationPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Application(ctx context.Context, id string) (*model.RecruitApplication, error) {
	path := "/api/v1/applications/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitApplication
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Candidates(ctx context.Context, page *int, limit *int, status *string, source *string, experienceLevel *string, skills []string, minScore *float64, search *string) (*model.RecruitCandidatePage, error) {
	path := "/api/v1/candidates"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if source != nil {
		q.Set("source", *source)
	}
	if experienceLevel != nil {
		q.Set("experience_level", *experienceLevel)
	}
	for _, e := range skills {
		q.Add("skills", e)
	}
	if minScore != nil {
		q.Set("min_score", strconv.FormatFloat(*minScore, 'f', -1, 64))
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out *model.RecruitCandidatePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Candidate(ctx context.Context, id string) (*model.RecruitCandidate, error) {
	path := "/api/v1/candidates/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitCandidate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PartnerStats(ctx context.Context, id string) (*model.RecruitPartnerStats, error) {
	path := "/api/v1/referrals/partners/" + url.PathEscape(id) + "/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitPartnerStats
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PartnerNetwork(ctx context.Context, id string) (*model.RecruitPartnerNetwork, error) {
	path := "/api/v1/referrals/partners/" + url.PathEscape(id) + "/network"
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitPartnerNetwork
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ReferralLeaderboard(ctx context.Context) ([]*model.RecruitPartnerStats, error) {
	path := "/api/v1/referrals/leaderboard"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RecruitPartnerStats
	err := c.rest.Get(ctx, path, q, h, "leaderboard", &out)
	return out, err
}

func (c *Client) Funnel(ctx context.Context) (*model.RecruitFunnel, error) {
	path := "/api/v1/analytics/funnel"
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitFunnel
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) FunnelByJob(ctx context.Context, jobID string) (*model.RecruitJobFunnel, error) {
	path := "/api/v1/analytics/funnel/" + url.PathEscape(jobID)
	q := url.Values{}
	h := http.Header{}
	var out *model.RecruitJobFunnel
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RecruiterPerformance(ctx context.Context) ([]*model.RecruitRecruiterRow, error) {
	path := "/api/v1/analytics/recruiter"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RecruitRecruiterRow
	err := c.rest.Get(ctx, path, q, h, "recruiters", &out)
	return out, err
}

func (c *Client) SLABreaches(ctx context.Context, breachedOnly *bool) (*model.RecruitSLABreaches, error) {
	path := "/api/v1/analytics/sla-breaches"
	q := url.Values{}
	if breachedOnly != nil {
		q.Set("breached_only", strconv.FormatBool(*breachedOnly))
	}
	h := http.Header{}
	var out *model.RecruitSLABreaches
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AiScoreDistribution(ctx context.Context) ([]*model.RecruitScoreRow, error) {
	path := "/api/v1/analytics/ai-scores"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RecruitScoreRow
	err := c.rest.Get(ctx, path, q, h, "score_distribution", &out)
	return out, err
}

func (c *Client) ReferralNetworkStats(ctx context.Context) ([]*model.RecruitNetworkRow, error) {
	path := "/api/v1/analytics/referral-network"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RecruitNetworkRow
	err := c.rest.Get(ctx, path, q, h, "network", &out)
	return out, err
}
