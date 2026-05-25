package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AnalyticsHandler struct {
	db *sqlx.DB
}

func NewAnalyticsHandler(db *sqlx.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/analytics")
	g.GET("/funnel",              h.RecruitmentFunnel)
	g.GET("/funnel/:job_id",      h.FunnelByJob)
	g.GET("/recruiter",           h.RecruiterPerformance)
	g.GET("/sla-breaches",        h.SLABreaches)
	g.GET("/ai-scores",           h.AIScoreDistribution)
	g.GET("/referral-network",    h.ReferralNetworkStats)
}

// GET /analytics/funnel
// Returns conversion funnel metrics for all active jobs
func (h *AnalyticsHandler) RecruitmentFunnel(c *gin.Context) {
	type FunnelRow struct {
		JobID            uuid.UUID `db:"job_id"             json:"job_id"`
		JobTitle         string    `db:"job_title"          json:"job_title"`
		JobCode          string    `db:"job_code"           json:"job_code"`
		TotalApplications int      `db:"total_applications" json:"total_applications"`
		CVReview         int       `db:"cv_review"          json:"cv_review"`
		PhoneScreen      int       `db:"phone_screen"       json:"phone_screen"`
		Technical        int       `db:"technical"          json:"technical"`
		FinalRound       int       `db:"final_round"        json:"final_round"`
		Offer            int       `db:"offer"              json:"offer"`
		Hired            int       `db:"hired"              json:"hired"`
		Rejected         int       `db:"rejected"           json:"rejected"`
		Withdrawn        int       `db:"withdrawn"          json:"withdrawn"`
		OfferRatePct     *float64  `db:"offer_rate_pct"     json:"offer_rate_pct"`
		AvgDaysToHire    *float64  `db:"avg_days_to_hire"   json:"avg_days_to_hire"`
	}

	var rows []FunnelRow
	if err := h.db.SelectContext(c.Request.Context(), &rows,
		`SELECT * FROM v_recruitment_funnel ORDER BY total_applications DESC`,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"funnel": rows, "total": len(rows)})
}

// GET /analytics/funnel/:job_id
func (h *AnalyticsHandler) FunnelByJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}
	var row struct {
		JobID            uuid.UUID `db:"job_id"             json:"job_id"`
		JobTitle         string    `db:"job_title"          json:"job_title"`
		TotalApplications int      `db:"total_applications" json:"total_applications"`
		CVReview         int       `db:"cv_review"          json:"cv_review"`
		PhoneScreen      int       `db:"phone_screen"       json:"phone_screen"`
		Technical        int       `db:"technical"          json:"technical"`
		FinalRound       int       `db:"final_round"        json:"final_round"`
		Offer            int       `db:"offer"              json:"offer"`
		Hired            int       `db:"hired"              json:"hired"`
		Rejected         int       `db:"rejected"           json:"rejected"`
		OfferRatePct     *float64  `db:"offer_rate_pct"     json:"offer_rate_pct"`
		AvgDaysToHire    *float64  `db:"avg_days_to_hire"   json:"avg_days_to_hire"`
	}
	if err := h.db.GetContext(c.Request.Context(), &row,
		`SELECT * FROM v_recruitment_funnel WHERE job_id = $1`, jobID,
	); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, row)
}

// GET /analytics/recruiter
func (h *AnalyticsHandler) RecruiterPerformance(c *gin.Context) {
	type RecruiterRow struct {
		RecruiterID        uuid.UUID `db:"recruiter_id"          json:"recruiter_id"`
		TotalManaged       int       `db:"total_managed"         json:"total_managed"`
		Hired              int       `db:"hired"                 json:"hired"`
		ActiveJobs         int       `db:"active_jobs"           json:"active_jobs"`
		AvgTimeToCloseDays *float64  `db:"avg_time_to_close_days" json:"avg_time_to_close_days"`
		StaleApplications  int       `db:"stale_applications"    json:"stale_applications"`
	}
	var rows []RecruiterRow
	if err := h.db.SelectContext(c.Request.Context(), &rows,
		`SELECT * FROM v_recruiter_performance ORDER BY hired DESC`,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recruiters": rows})
}

// GET /analytics/sla-breaches
func (h *AnalyticsHandler) SLABreaches(c *gin.Context) {
	type SLARow struct {
		ApplicationID uuid.UUID `db:"application_id" json:"application_id"`
		JobID         uuid.UUID `db:"job_id"         json:"job_id"`
		CandidateID   uuid.UUID `db:"candidate_id"   json:"candidate_id"`
		RecruiterID   uuid.UUID `db:"recruiter_id"   json:"recruiter_id"`
		Status        string    `db:"status"         json:"status"`
		DaysInStage   int       `db:"days_in_stage"  json:"days_in_stage"`
		SLADays       int       `db:"sla_days"       json:"sla_days"`
		IsBreached    bool      `db:"is_breached"    json:"is_breached"`
	}

	onlyBreached := c.Query("breached_only") == "true"
	query := `SELECT * FROM v_sla_breaches ORDER BY days_in_stage DESC`
	if onlyBreached {
		query = `SELECT * FROM v_sla_breaches WHERE is_breached = true ORDER BY days_in_stage DESC`
	}

	var rows []SLARow
	if err := h.db.SelectContext(c.Request.Context(), &rows, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	breachedCount := 0
	for _, r := range rows {
		if r.IsBreached {
			breachedCount++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"applications":   rows,
		"total":          len(rows),
		"breached_count": breachedCount,
	})
}

// GET /analytics/ai-scores
func (h *AnalyticsHandler) AIScoreDistribution(c *gin.Context) {
	type ScoreRow struct {
		JobID              uuid.UUID `db:"job_id"               json:"job_id"`
		JobTitle           string    `db:"job_title"            json:"job_title"`
		ScoredApplications int       `db:"scored_applications"  json:"scored_applications"`
		AvgScore           *float64  `db:"avg_score"            json:"avg_score"`
		MinScore           *float64  `db:"min_score"            json:"min_score"`
		MaxScore           *float64  `db:"max_score"            json:"max_score"`
		HighMatch          int       `db:"high_match"           json:"high_match"`
		MediumMatch        int       `db:"medium_match"         json:"medium_match"`
		LowMatch           int       `db:"low_match"            json:"low_match"`
	}
	var rows []ScoreRow
	if err := h.db.SelectContext(c.Request.Context(), &rows,
		`SELECT * FROM v_ai_score_distribution ORDER BY avg_score DESC`,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"score_distribution": rows})
}

// GET /analytics/referral-network
func (h *AnalyticsHandler) ReferralNetworkStats(c *gin.Context) {
	type NetworkRow struct {
		PartnerID      uuid.UUID `db:"partner_id"     json:"partner_id"`
		FullName       string    `db:"full_name"      json:"full_name"`
		NetworkLevel   int       `db:"network_level"  json:"network_level"`
		HiredReferrals int       `db:"hired_referrals" json:"hired_referrals"`
		DirectDownline int       `db:"direct_downline" json:"direct_downline"`
	}
	var rows []NetworkRow
	if err := h.db.SelectContext(c.Request.Context(), &rows,
		`SELECT * FROM v_referral_network_stats ORDER BY hired_referrals DESC LIMIT 100`,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"network": rows})
}
