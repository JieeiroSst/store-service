package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(NewHandler),
)

type Handler struct {
	svc ports.BonusService
	log *zap.Logger
}

func NewHandler(svc ports.BonusService, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log.Named("http-handler")}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/bonus")
	v1.POST("/rewards", h.RecordReward)
	v1.GET("/users/:user_id/rewards", h.ListUserRewards)
	v1.GET("/users/:user_id/balances", h.GetUserBalances)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

type recordRewardRequest struct {
	EventID     string  `json:"event_id" binding:"required"`
	UserID      string  `json:"user_id" binding:"required"`
	RefCode     string  `json:"ref_code"`
	Source      string  `json:"source"`
	RewardType  string  `json:"reward_type" binding:"required"`
	RewardValue float64 `json:"reward_value" binding:"required"`
}

// RecordReward POST /api/v1/bonus/rewards — 201 when recorded, 200 for a duplicate event_id.
func (h *Handler) RecordReward(c *gin.Context) {
	var req recordRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorJSON(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	reward, created, err := h.svc.RecordReward(c.Request.Context(), ports.RecordRewardRequest{
		EventID:    req.EventID,
		UserID:     req.UserID,
		RefCode:    req.RefCode,
		Source:     req.Source,
		RewardType: domain.RewardType(req.RewardType),
		Value:      req.RewardValue,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	if !created {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"duplicate": true}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": reward})
}

// ListUserRewards GET /api/v1/bonus/users/:user_id/rewards?limit=&offset=
func (h *Handler) ListUserRewards(c *gin.Context) {
	limit := queryInt(c, "limit", 20, 1, 100)
	offset := queryInt(c, "offset", 0, 0, 1<<31-1)
	rewards, err := h.svc.ListUserRewards(c.Request.Context(), c.Param("user_id"), limit, offset)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rewards, "meta": gin.H{"count": len(rewards)}})
}

// GetUserBalances GET /api/v1/bonus/users/:user_id/balances
func (h *Handler) GetUserBalances(c *gin.Context) {
	balances, err := h.svc.GetUserBalances(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": balances})
}

func (h *Handler) fail(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrInvalidReward) {
		errorJSON(c, http.StatusUnprocessableEntity, "INVALID_REWARD", err.Error())
		return
	}
	h.log.Error("request failed", zap.Error(err), zap.String("path", c.Request.URL.Path))
	errorJSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
}

func errorJSON(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}

func queryInt(c *gin.Context, key string, def, min, max int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < min {
		return def
	}
	if v > max {
		return max
	}
	return v
}
