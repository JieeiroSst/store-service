package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AcceptAssignment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	assignment, err := h.assignments.AcceptAssignment(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignment)
}

type rejectAssignmentRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) RejectAssignment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req rejectAssignmentRequest
	_ = c.ShouldBindJSON(&req)

	assignment, err := h.assignments.RejectAssignment(c.Request.Context(), id, req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *Handler) CompleteAssignment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	assignment, err := h.assignments.CompleteAssignment(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *Handler) ListAssignmentsByDriver(c *gin.Context) {
	driverID := c.Query("driver_id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id query param is required"})
		return
	}
	assignments, err := h.assignments.ListAssignmentsByDriver(c.Request.Context(), driverID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignments)
}
