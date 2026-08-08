package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type ContentHandler struct {
	usecase port.ContentUsecase
}

func NewContentHandler(usecase port.ContentUsecase) *ContentHandler {
	return &ContentHandler{usecase: usecase}
}

type mediaRequest struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Thumbnail string `json:"thumbnail"`
}

type createPostRequest struct {
	SaveAsDraft     bool           `json:"save_as_draft"`
	Title           string         `json:"title"`
	Text            string         `json:"text"`
	Hashtags        []string       `json:"hashtags"`
	Media           []mediaRequest `json:"media"`
	MediaKind       string         `json:"media_kind"`
	Channels        []string       `json:"channels"`
	ScheduledAt     *time.Time     `json:"scheduled_at"`
	Timezone        string         `json:"timezone"`
	CronExpr        string         `json:"cron_expr"`
	MaxRunsPerDay   int            `json:"max_runs_per_day"`
	MaxRunsPerMonth int            `json:"max_runs_per_month"`
	Campaign        string         `json:"campaign"`
	CreatedBy       string         `json:"created_by"`
}

func toModelChannels(in []string) []model.Channel {
	out := make([]model.Channel, 0, len(in))
	for _, ch := range in {
		out = append(out, model.Channel(ch))
	}
	return out
}

func toModelMedia(in []mediaRequest) []model.Media {
	out := make([]model.Media, 0, len(in))
	for _, m := range in {
		out = append(out, model.Media{URL: m.URL, Type: model.MediaType(m.Type), Thumbnail: m.Thumbnail})
	}
	return out
}

func (h *ContentHandler) CreatePost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.usecase.CreatePost(c.Request.Context(), port.CreatePostInput{
		SaveAsDraft:     req.SaveAsDraft,
		Title:           req.Title,
		Text:            req.Text,
		Hashtags:        req.Hashtags,
		Media:           toModelMedia(req.Media),
		MediaKind:       model.MediaKind(req.MediaKind),
		Channels:        toModelChannels(req.Channels),
		ScheduledAt:     req.ScheduledAt,
		Timezone:        req.Timezone,
		CronExpr:        req.CronExpr,
		MaxRunsPerDay:   req.MaxRunsPerDay,
		MaxRunsPerMonth: req.MaxRunsPerMonth,
		Campaign:        req.Campaign,
		CreatedBy:       req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

type updatePostRequest struct {
	Title       *string        `json:"title"`
	Text        *string        `json:"text"`
	Hashtags    []string       `json:"hashtags"`
	Media       []mediaRequest `json:"media"`
	MediaKind   *string        `json:"media_kind"`
	Channels    []string       `json:"channels"`
	ScheduledAt *time.Time     `json:"scheduled_at"`
	Timezone    *string        `json:"timezone"`
	CronExpr    *string        `json:"cron_expr"`

	MaxRunsPerDay   *int    `json:"max_runs_per_day"`
	MaxRunsPerMonth *int    `json:"max_runs_per_month"`
	Campaign        *string `json:"campaign"`
}

func (h *ContentHandler) UpdatePost(c *gin.Context) {
	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := port.UpdatePostInput{
		Title:           req.Title,
		Text:            req.Text,
		Hashtags:        req.Hashtags,
		ScheduledAt:     req.ScheduledAt,
		Timezone:        req.Timezone,
		CronExpr:        req.CronExpr,
		MaxRunsPerDay:   req.MaxRunsPerDay,
		MaxRunsPerMonth: req.MaxRunsPerMonth,
		Campaign:        req.Campaign,
	}
	if req.Media != nil {
		input.Media = toModelMedia(req.Media)
	}
	if req.MediaKind != nil {
		mk := model.MediaKind(*req.MediaKind)
		input.MediaKind = &mk
	}
	if req.Channels != nil {
		input.Channels = toModelChannels(req.Channels)
	}

	post, err := h.usecase.UpdatePost(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *ContentHandler) GetPost(c *gin.Context) {
	post, err := h.usecase.GetPost(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *ContentHandler) ListPosts(c *gin.Context) {
	limit := queryInt32(c, "limit", 20)
	offset := queryInt32(c, "offset", 0)
	status := model.PostStatus(c.Query("status"))
	campaign := c.Query("campaign")

	posts, err := h.usecase.ListPosts(c.Request.Context(), limit, offset, status, campaign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

type actorRequest struct {
	ChangedBy string `json:"changed_by"`
}

func bindActor(c *gin.Context) string {
	var req actorRequest
	_ = c.ShouldBindJSON(&req)
	return req.ChangedBy
}

func (h *ContentHandler) SubmitForReview(c *gin.Context) {
	post, err := h.usecase.SubmitForReview(c.Request.Context(), c.Param("id"), bindActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *ContentHandler) ApprovePost(c *gin.Context) {
	post, err := h.usecase.ApprovePost(c.Request.Context(), c.Param("id"), bindActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

type rejectPostRequest struct {
	ChangedBy string `json:"changed_by"`
	Reason    string `json:"reason"`
}

func (h *ContentHandler) RejectPost(c *gin.Context) {
	var req rejectPostRequest
	_ = c.ShouldBindJSON(&req)

	post, err := h.usecase.RejectPost(c.Request.Context(), c.Param("id"), req.ChangedBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *ContentHandler) PublishNow(c *gin.Context) {
	post, err := h.usecase.PublishNow(c.Request.Context(), c.Param("id"), bindActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *ContentHandler) CancelPost(c *gin.Context) {
	if err := h.usecase.CancelScheduledPost(c.Request.Context(), c.Param("id"), bindActor(c)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func queryInt32(c *gin.Context, key string, fallback int32) int32 {
	v := c.Query(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return int32(n)
}
