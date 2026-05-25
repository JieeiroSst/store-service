package handler

import (
	"net/http"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CandidateHandler struct {
	svc port.CandidateService
}

func NewCandidateHandler(svc port.CandidateService) *CandidateHandler {
	return &CandidateHandler{svc: svc}
}

func (h *CandidateHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/candidates")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PUT("/:id/status", h.Transition)
		g.POST("/:id/enrich", h.EnrichWithAI)
	}
}

// POST /candidates
func (h *CandidateHandler) Create(c *gin.Context) {
	var req struct {
		FullName      string `json:"full_name"      binding:"required"`
		Email         string `json:"email"          binding:"required,email"`
		Phone         string `json:"phone"`
		Source        string `json:"source"         binding:"required"`
		ResumeURL     string `json:"resume_url"`
		ReferralToken string `json:"referral_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Create(c.Request.Context(), port.CreateCandidateCommand{
		FullName:      req.FullName,
		Email:         req.Email,
		Phone:         req.Phone,
		Source:        candidate.SourceChannel(req.Source),
		ResumeURL:     req.ResumeURL,
		ReferralToken: req.ReferralToken,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GET /candidates
func (h *CandidateHandler) List(c *gin.Context) {
	var filter candidate.Filter
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

// GET /candidates/:id
func (h *CandidateHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PATCH /candidates/:id
func (h *CandidateHandler) Update(c *gin.Context) {
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
	result, err := h.svc.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PUT /candidates/:id/status
func (h *CandidateHandler) Transition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Transition(c.Request.Context(), id, candidate.Status(req.Status), req.Reason); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// POST /candidates/:id/enrich
func (h *CandidateHandler) EnrichWithAI(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.EnrichWithAI(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "enrichment complete"})
}
