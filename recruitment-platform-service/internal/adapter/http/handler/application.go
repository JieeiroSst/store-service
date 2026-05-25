package handler

import (
	"net/http"
	"time"

	domainapp "github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApplicationHandler struct {
	svc port.ApplicationService
}

func NewApplicationHandler(svc port.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

func (h *ApplicationHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/applications")
	{
		g.POST("", h.Apply)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id/stage", h.MoveStage)
		g.POST("/:id/interviews", h.ScheduleInterview)
		g.POST("/:id/interviews/:interviewId/feedback", h.SubmitFeedback)
		g.POST("/:id/offer", h.ExtendOffer)
		g.POST("/:id/reject", h.Reject)
	}
}

// POST /applications
func (h *ApplicationHandler) Apply(c *gin.Context) {
	var req struct {
		JobID               string  `json:"job_id"                 binding:"required,uuid"`
		CandidateID         string  `json:"candidate_id"           binding:"required,uuid"`
		ReferredByPartnerID *string `json:"referred_by_partner_id"`
		CoverLetter         string  `json:"cover_letter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recruiterID := mustGetUserID(c)
	cmd := port.ApplyCommand{
		JobID:       uuid.MustParse(req.JobID),
		CandidateID: uuid.MustParse(req.CandidateID),
		RecruiterID: recruiterID,
		CoverLetter: req.CoverLetter,
	}
	if req.ReferredByPartnerID != nil {
		pid := uuid.MustParse(*req.ReferredByPartnerID)
		cmd.ReferredByPartnerID = &pid
	}

	app, err := h.svc.Apply(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, app)
}

// GET /applications
func (h *ApplicationHandler) List(c *gin.Context) {
	var filter domainapp.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	result, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /applications/:id
func (h *ApplicationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	app, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

// PUT /applications/:id/stage
func (h *ApplicationHandler) MoveStage(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		StageID string `json:"stage_id" binding:"required,uuid"`
		Status  string `json:"status"   binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.svc.MoveStage(c.Request.Context(), port.MoveStageCommand{
		ApplicationID: id,
		StageID:       uuid.MustParse(req.StageID),
		Status:        req.Status,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

// POST /applications/:id/interviews
func (h *ApplicationHandler) ScheduleInterview(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		Round          int      `json:"round"           binding:"required"`
		Title          string   `json:"title"           binding:"required"`
		InterviewerIDs []string `json:"interviewer_ids" binding:"required"`
		ScheduledAt    string   `json:"scheduled_at"    binding:"required"`
		DurationMin    int      `json:"duration_min"`
		MeetingURL     string   `json:"meeting_url"`
		Type           string   `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduled_at format, use RFC3339"})
		return
	}
	interviewerIDs := make([]uuid.UUID, 0, len(req.InterviewerIDs))
	for _, idStr := range req.InterviewerIDs {
		interviewerIDs = append(interviewerIDs, uuid.MustParse(idStr))
	}
	app, err := h.svc.ScheduleInterview(c.Request.Context(), port.ScheduleInterviewCommand{
		ApplicationID:  id,
		Round:          req.Round,
		Title:          req.Title,
		InterviewerIDs: interviewerIDs,
		ScheduledAt:    scheduledAt,
		DurationMin:    req.DurationMin,
		MeetingURL:     req.MeetingURL,
		Type:           req.Type,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, app)
}

// POST /applications/:id/interviews/:interviewId/feedback
func (h *ApplicationHandler) SubmitFeedback(c *gin.Context) {
	appID, _ := uuid.Parse(c.Param("id"))
	ivID, _ := uuid.Parse(c.Param("interviewId"))
	var req struct {
		Decision   string `json:"decision"   binding:"required,oneof=pass fail hold"`
		Score      int    `json:"score"      binding:"required,min=1,max=5"`
		Strengths  string `json:"strengths"`
		Weaknesses string `json:"weaknesses"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SubmitFeedback(c.Request.Context(), port.SubmitFeedbackCommand{
		ApplicationID: appID,
		InterviewID:   ivID,
		SubmittedBy:   mustGetUserID(c),
		Decision:      req.Decision,
		Score:         req.Score,
		Strengths:     req.Strengths,
		Weaknesses:    req.Weaknesses,
		Notes:         req.Notes,
	}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "feedback submitted"})
}

// POST /applications/:id/offer
func (h *ApplicationHandler) ExtendOffer(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		SalaryAmount   int64    `json:"salary_amount"   binding:"required"`
		SalaryCurrency string   `json:"salary_currency" binding:"required"`
		StartDate      string   `json:"start_date"      binding:"required"`
		Title          string   `json:"title"           binding:"required"`
		Benefits       []string `json:"benefits"`
		ExpiresAt      string   `json:"expires_at"      binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	expiresAt, _ := time.Parse(time.RFC3339, req.ExpiresAt)

	app, err := h.svc.ExtendOffer(c.Request.Context(), port.ExtendOfferCommand{
		ApplicationID: id,
		Salary:        shared.Money{Amount: req.SalaryAmount, Currency: req.SalaryCurrency},
		StartDate:     startDate,
		Title:         req.Title,
		Benefits:      req.Benefits,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

// POST /applications/:id/reject
func (h *ApplicationHandler) Reject(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		Reason string `json:"reason" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Reject(c.Request.Context(), port.RejectCommand{
		ApplicationID: id,
		Reason:        req.Reason,
		Note:          req.Note,
	}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "application rejected"})
}
