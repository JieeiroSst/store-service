package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type restrictionRequest struct {
	RestrictionType    string `json:"restriction_type" binding:"required"`
	RestrictedEntityID int64  `json:"restricted_entity_id" binding:"required"`
	IsExclude          bool   `json:"is_exclude"`
}

func (h *Handler) CreateRestriction(c *gin.Context) {
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}
	var req restrictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	restriction, err := h.restrictions.CreateRestriction(c.Request.Context(), &model.CouponRestriction{
		CouponID:           couponID,
		RestrictionType:    model.RestrictionType(req.RestrictionType),
		RestrictedEntityID: req.RestrictedEntityID,
		IsExclude:          req.IsExclude,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, restriction)
}

func (h *Handler) ListRestrictions(c *gin.Context) {
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}
	restrictions, err := h.restrictions.ListRestrictionsByCoupon(c.Request.Context(), couponID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, restrictions)
}

func (h *Handler) UpdateRestriction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.restrictions.GetRestriction(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	var req restrictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.RestrictionType = model.RestrictionType(req.RestrictionType)
	existing.RestrictedEntityID = req.RestrictedEntityID
	existing.IsExclude = req.IsExclude

	restriction, err := h.restrictions.UpdateRestriction(c.Request.Context(), existing)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, restriction)
}

func (h *Handler) DeleteRestriction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.restrictions.DeleteRestriction(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
