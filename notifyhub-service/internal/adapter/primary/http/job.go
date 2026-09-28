package http

import (
	"net/http"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateJob(c *gin.Context) {
	var j model.NotifyJob
	if err := c.ShouldBindJSON(&j); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.jobs.Create(c.Request.Context(), &j)
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListJobs(c *gin.Context) {
	page, size := paginate(c)
	jobs, total, err := h.jobs.List(c.Request.Context(), port.JobFilter{
		Status:   c.Query("status"),
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	respondPage(c, jobs, total, page, size)
}

func (h *Handler) GetJob(c *gin.Context) {
	j, err := h.jobs.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, j)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	j, err := h.jobs.Update(c.Request.Context(), c.Param("id"), bindTo[model.NotifyJob](c))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, j)
}

func (h *Handler) PauseJob(c *gin.Context) {
	if err := h.jobs.Pause(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"status": model.JobStatusPaused})
}

func (h *Handler) ResumeJob(c *gin.Context) {
	if err := h.jobs.Resume(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"status": model.JobStatusActive})
}

func (h *Handler) TriggerJob(c *gin.Context) {
	if err := h.jobs.Trigger(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"triggered": true, "job_id": c.Param("id")})
}

func (h *Handler) DeleteJob(c *gin.Context) {
	if err := h.jobs.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) ListHistory(c *gin.Context) {
	page, size := paginate(c)
	hist, total, err := h.history.List(c.Request.Context(), port.HistoryFilter{
		JobID:    c.Query("job_id"),
		Status:   c.Query("status"),
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	respondPage(c, hist, total, page, size)
}
