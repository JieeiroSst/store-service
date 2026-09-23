package http

import (
	"net/http"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type profileRequest struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

func (h *Handler) GetProfile(c *gin.Context) {
	p, err := h.social.GetProfile(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req profileRequest
	if !bind(c, &req) {
		return
	}
	p, err := h.social.UpdateProfile(c.Request.Context(), port.UpdateProfileInput{
		UserID: c.Param("userId"), Username: req.Username, Bio: req.Bio, AvatarURL: req.AvatarURL,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type commentRequest struct {
	UserID   string `json:"user_id"`
	Body     string `json:"body" binding:"required"`
	ParentID int64  `json:"parent_id"`
}

type userRequest struct {
	UserID string `json:"user_id"`
}

func (h *Handler) AddComment(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req commentRequest
	if !bind(c, &req) {
		return
	}
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	comment, err := h.social.AddComment(c.Request.Context(), eventID, user, req.Body, req.ParentID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// ListComments: ?sort=newest|likes&limit=&cursor= -> {items, next_cursor, is_last_page}
func (h *Handler) ListComments(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	page, err := h.social.ListComments(c.Request.Context(), eventID, c.Query("sort"), queryInt(c, "limit"), c.Query("cursor"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *Handler) DeleteComment(c *gin.Context) {
	id, ok := pathInt(c, "commentId")
	if !ok {
		return
	}
	user, ok := h.caller(c, c.Query("user_id"))
	if !ok {
		return
	}
	if err := h.social.DeleteComment(c.Request.Context(), id, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) LikeComment(c *gin.Context) {
	id, ok := pathInt(c, "commentId")
	if !ok {
		return
	}
	var req userRequest
	_ = c.ShouldBindJSON(&req)
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	if err := h.social.LikeComment(c.Request.Context(), id, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UnlikeComment(c *gin.Context) {
	id, ok := pathInt(c, "commentId")
	if !ok {
		return
	}
	user, ok := h.caller(c, c.Query("user_id"))
	if !ok {
		return
	}
	if err := h.social.UnlikeComment(c.Request.Context(), id, user); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Watch(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req userRequest
	_ = c.ShouldBindJSON(&req)
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	if err := h.social.Watch(c.Request.Context(), user, eventID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Unwatch(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	user, ok := h.caller(c, c.Query("user_id"))
	if !ok {
		return
	}
	if err := h.social.Unwatch(c.Request.Context(), user, eventID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Watchlist(c *gin.Context) {
	events, err := h.social.Watchlist(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventViews(events))
}

// Leaderboard: ?metric=profit|volume&window=day|week|month|all&limit=
func (h *Handler) Leaderboard(c *gin.Context) {
	rows, err := h.social.Leaderboard(c.Request.Context(), c.Query("metric"), c.Query("window"), queryInt(c, "limit"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}
