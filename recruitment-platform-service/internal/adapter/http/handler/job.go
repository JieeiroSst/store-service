package handler

import (
	"net/http"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobHandler struct {
	svc port.JobService
}

func NewJobHandler(svc port.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

func (h *JobHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/jobs")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PUT("/:id/publish", h.Publish)
		g.PUT("/:id/pause", h.Pause)
		g.PUT("/:id/close", h.Close)
		g.GET("/:id/recommended-candidates", h.RecommendedCandidates)
	}
}

// POST /jobs
func (h *JobHandler) Create(c *gin.Context) {
	var req struct {
		Title           string   `json:"title"             binding:"required"`
		Code            string   `json:"code"              binding:"required"`
		DepartmentID    string   `json:"department_id"     binding:"required,uuid"`
		HiringManagerID string   `json:"hiring_manager_id" binding:"required,uuid"`
		RecruiterIDs    []string `json:"recruiter_ids"`
		Description     string   `json:"description"`
		Requirements    []string `json:"requirements"`
		Skills          []string `json:"skills"`
		JobType         string   `json:"job_type"`
		WorkMode        string   `json:"work_mode"`
		City            string   `json:"city"`
		Country         string   `json:"country"`
		SalaryMinAmount *int64   `json:"salary_min_amount"`
		SalaryMaxAmount *int64   `json:"salary_max_amount"`
		SalaryCurrency  string   `json:"salary_currency"`
		Headcount       int      `json:"headcount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recruiterIDs := make([]uuid.UUID, 0, len(req.RecruiterIDs))
	for _, rid := range req.RecruiterIDs {
		recruiterIDs = append(recruiterIDs, uuid.MustParse(rid))
	}

	cmd := port.CreateJobCommand{
		Title:           req.Title,
		Code:            req.Code,
		DepartmentID:    uuid.MustParse(req.DepartmentID),
		HiringManagerID: uuid.MustParse(req.HiringManagerID),
		RecruiterIDs:    recruiterIDs,
		Description:     req.Description,
		Requirements:    req.Requirements,
		Skills:          req.Skills,
		JobType:         job.JobType(req.JobType),
		WorkMode:        job.WorkMode(req.WorkMode),
		Location:        shared.Address{City: req.City, Country: req.Country},
		Headcount:       req.Headcount,
	}
	if req.SalaryMinAmount != nil {
		cmd.SalaryMin = &shared.Money{Amount: *req.SalaryMinAmount, Currency: req.SalaryCurrency}
	}
	if req.SalaryMaxAmount != nil {
		cmd.SalaryMax = &shared.Money{Amount: *req.SalaryMaxAmount, Currency: req.SalaryCurrency}
	}

	j, err := h.svc.Create(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, j)
}

// GET /jobs
func (h *JobHandler) List(c *gin.Context) {
	var filter job.Filter
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

// GET /jobs/:id
func (h *JobHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	j, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}

// PATCH /jobs/:id
func (h *JobHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	j, err := h.svc.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, j)
}

// PUT /jobs/:id/publish
func (h *JobHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Publish(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job published"})
}

// PUT /jobs/:id/pause
func (h *JobHandler) Pause(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Pause(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job paused"})
}

// PUT /jobs/:id/close
func (h *JobHandler) Close(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.Close(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job closed"})
}

// GET /jobs/:id/recommended-candidates
func (h *JobHandler) RecommendedCandidates(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit := 10
	candidates, err := h.svc.GetRecommendedCandidates(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}
