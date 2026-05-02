package handler

import (
	"net/http"
	"strconv"
	"time"

	jobSvc "github.com/JIeeiroSst/notifyhub-service/internal/job"
	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/pkg/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/store/mysql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	store      *mysql.Store
	jobSvc     *jobSvc.Service
	channelSvc *jobSvc.ChannelService
	dsSvc      *jobSvc.DataSourceService
	tmplSvc    *jobSvc.TemplateService
	logger     *zap.Logger
}

func New(
	store *mysql.Store,
	js *jobSvc.Service,
	cs *jobSvc.ChannelService,
	ds *jobSvc.DataSourceService,
	ts *jobSvc.TemplateService,
	log *zap.Logger,
) *Handler {
	return &Handler{
		store:      store,
		jobSvc:     js,
		channelSvc: cs,
		dsSvc:      ds,
		tmplSvc:    ts,
		logger:     log,
	}
}


func respond(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{
		"data":       data,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString("request_id"),
	})
}

func respondErr(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"error":      msg,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString("request_id"),
	})
}

func paginate(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}


func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"ts":      time.Now().Unix(),
		"version": "1.0.0",
	})
}


func (h *Handler) CreateChannel(c *gin.Context) {
	var ch model.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.channelSvc.Create(c.Request.Context(), &ch)
	if err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListChannels(c *gin.Context) {
	ct := c.Query("type")
	var active *bool
	if v := c.Query("active"); v != "" {
		b := v == "true"
		active = &b
	}
	channels, err := h.store.ListChannels(c.Request.Context(), ct, active)
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, channels)
}

func (h *Handler) GetChannel(c *gin.Context) {
	ch, err := h.store.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "channel not found")
		return
	}
	respond(c, http.StatusOK, ch)
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	ch, err := h.store.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "channel not found")
		return
	}
	if err := c.ShouldBindJSON(ch); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	ch.ID = c.Param("id")
	ch.UpdatedAt = time.Now()
	if err := h.store.UpdateChannel(c.Request.Context(), ch); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, ch)
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	if err := h.store.DeleteChannel(c.Request.Context(), c.Param("id")); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) CreateDataSource(c *gin.Context) {
	var ds model.DataSource
	if err := c.ShouldBindJSON(&ds); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.dsSvc.Create(c.Request.Context(), &ds)
	if err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListDataSources(c *gin.Context) {
	dss, err := h.store.ListDataSources(c.Request.Context())
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, dss)
}

func (h *Handler) GetDataSource(c *gin.Context) {
	ds, err := h.store.GetDataSource(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "data source not found")
		return
	}
	respond(c, http.StatusOK, ds)
}

func (h *Handler) UpdateDataSource(c *gin.Context) {
	ds, err := h.store.GetDataSource(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "data source not found")
		return
	}
	if err := c.ShouldBindJSON(ds); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	ds.ID = c.Param("id")
	ds.UpdatedAt = time.Now()
	if err := h.store.UpdateDataSource(c.Request.Context(), ds); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, ds)
}

func (h *Handler) DeleteDataSource(c *gin.Context) {
	ds, err := h.store.GetDataSource(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "data source not found")
		return
	}
	ds.IsActive = false
	ds.UpdatedAt = time.Now()
	if err := h.store.UpdateDataSource(c.Request.Context(), ds); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}


func (h *Handler) CreateTemplate(c *gin.Context) {
	var t model.Template
	if err := c.ShouldBindJSON(&t); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.tmplSvc.Create(c.Request.Context(), &t)
	if err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	ts, err := h.store.ListTemplates(c.Request.Context(), c.Query("channel"))
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, ts)
}

func (h *Handler) GetTemplate(c *gin.Context) {
	t, err := h.store.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "template not found")
		return
	}
	respond(c, http.StatusOK, t)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	var patch model.Template
	if err := c.ShouldBindJSON(&patch); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.tmplSvc.Update(c.Request.Context(), c.Param("id"), &patch)
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, updated)
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	t, err := h.store.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "template not found")
		return
	}
	t.IsActive = false
	t.UpdatedAt = time.Now()
	if err := h.store.UpdateTemplate(c.Request.Context(), t); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) CreateJob(c *gin.Context) {
	var j model.NotifyJob
	if err := c.ShouldBindJSON(&j); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.jobSvc.CreateJob(c.Request.Context(), &j)
	if err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	metrics.JobExecutionsTotal.WithLabelValues(string(created.ScheduleType), "created").Inc()
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListJobs(c *gin.Context) {
	page, size := paginate(c)
	jobs, total, err := h.store.ListJobs(c.Request.Context(), c.Query("status"), page, size)
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       jobs,
		"total":      total,
		"page":       page,
		"page_size":  size,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString("request_id"),
	})
}

func (h *Handler) GetJob(c *gin.Context) {
	j, err := h.store.GetJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondErr(c, http.StatusNotFound, "job not found")
		return
	}
	respond(c, http.StatusOK, j)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	var patch model.NotifyJob
	if err := c.ShouldBindJSON(&patch); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.jobSvc.UpdateJob(c.Request.Context(), c.Param("id"), &patch)
	if err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	respond(c, http.StatusOK, updated)
}

func (h *Handler) PauseJob(c *gin.Context) {
	if err := h.jobSvc.PauseJob(c.Request.Context(), c.Param("id")); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"status": "paused"})
}

func (h *Handler) ResumeJob(c *gin.Context) {
	if err := h.jobSvc.ResumeJob(c.Request.Context(), c.Param("id")); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"status": "active"})
}

func (h *Handler) TriggerJob(c *gin.Context) {
	if err := h.jobSvc.TriggerNow(c.Request.Context(), c.Param("id")); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"triggered": true, "job_id": c.Param("id")})
}

func (h *Handler) DeleteJob(c *gin.Context) {
	if err := h.jobSvc.DeleteJob(c.Request.Context(), c.Param("id")); err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) ListHistory(c *gin.Context) {
	page, size := paginate(c)
	hist, total, err := h.store.ListHistory(
		c.Request.Context(),
		c.Query("job_id"),
		c.Query("status"),
		page, size,
	)
	if err != nil {
		respondErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       hist,
		"total":      total,
		"page":       page,
		"page_size":  size,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString("request_id"),
	})
}

func (h *Handler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	hist, _, err := h.store.ListHistory(c.Request.Context(), "", "", 1, 1)
	if err != nil || len(hist) == 0 {
		respondErr(c, http.StatusNotFound, "history record not found")
		return
	}
	_ = id
	respond(c, http.StatusOK, hist[0])
}

type schedInfo interface {
	ListScheduled() []string
}

func SchedulerStatus(sched schedInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		ids := sched.ListScheduled()
		c.JSON(http.StatusOK, gin.H{
			"scheduled_jobs": ids,
			"count":          len(ids),
			"ts":             time.Now().Unix(),
		})
	}
}

// requestID generates and stores a unique request ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
