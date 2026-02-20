package handler

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"github.com/JIeeiroSst/qr-service/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type QRHandler struct {
	qrSvc   port.QRCodeService
	scanSvc port.ScanHistoryService
	logger  *zap.Logger
}

func NewQRHandler(
	qrSvc port.QRCodeService,
	scanSvc port.ScanHistoryService,
	logger *zap.Logger,
) *QRHandler {
	return &QRHandler{
		qrSvc:   qrSvc,
		scanSvc: scanSvc,
		logger:  logger,
	}
}

// Generate godoc
// @Summary     Generate a new QR code
// @Tags        qr-codes
// @Accept      json
// @Produce     json
// @Param       body body port.GenerateQRRequest true "QR generation request"
// @Success     201 {object} response.APIResponse
// @Router      /api/v1/qr [post]
func (h *QRHandler) Generate(c *gin.Context) {
	var req port.GenerateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.qrSvc.Generate(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("generate qr failed", zap.Error(err))
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

// GetByID godoc
// @Summary     Get a QR code by ID
// @Tags        qr-codes
// @Produce     json
// @Param       id path string true "QR code ID"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id} [get]
func (h *QRHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	qr, err := h.qrSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, qr)
}

// List godoc
// @Summary     List QR codes with pagination and filters
// @Tags        qr-codes
// @Produce     json
// @Param       page       query int    false "Page number"
// @Param       limit      query int    false "Items per page"
// @Param       status     query string false "Filter by status"
// @Param       type       query string false "Filter by type"
// @Param       search     query string false "Search by title/content"
// @Param       created_by query string false "Filter by creator"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr [get]
func (h *QRHandler) List(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	filter := port.QRCodeFilter{
		Status:    c.Query("status"),
		Type:      c.Query("type"),
		Search:    c.Query("search"),
		CreatedBy: c.Query("created_by"),
		Page:      page,
		Limit:     limit,
	}

	result, err := h.qrSvc.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Update godoc
// @Summary     Update QR code metadata
// @Tags        qr-codes
// @Accept      json
// @Produce     json
// @Param       id   path string               true "QR code ID"
// @Param       body body port.UpdateQRRequest true "Update request"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id} [put]
func (h *QRHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req port.UpdateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.qrSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, updated)
}

// UpdateContent godoc
// @Summary     Dynamically update the target content/URL of a QR code
// @Description The QR image stays the same but points to new content
// @Tags        qr-codes
// @Accept      json
// @Produce     json
// @Param       id   path string                      true "QR code ID"
// @Param       body body port.UpdateContentRequest   true "New content"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id}/content [patch]
func (h *QRHandler) UpdateContent(c *gin.Context) {
	id := c.Param("id")
	var req port.UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.qrSvc.UpdateContent(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, updated)
}

// Delete godoc
// @Summary     Delete a QR code and its scan history
// @Tags        qr-codes
// @Produce     json
// @Param       id path string true "QR code ID"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id} [delete]
func (h *QRHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.qrSvc.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "deleted successfully"})
}

// Redirect godoc
// @Summary     Scan / redirect a QR code short link
// @Tags        qr-codes
// @Produce     json
// @Param       shortCode path string true "Short code"
// @Success     302
// @Router      /qr/scan/{shortCode} [get]
func (h *QRHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	meta := &port.ScanMeta{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	}

	redirectURL, err := h.qrSvc.Redirect(c.Request.Context(), shortCode, meta)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// Regenerate godoc
// @Summary     Regenerate the QR code image for an existing entry
// @Tags        qr-codes
// @Produce     json
// @Param       id path string true "QR code ID"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id}/regenerate [post]
func (h *QRHandler) Regenerate(c *gin.Context) {
	id := c.Param("id")
	result, err := h.qrSvc.Regenerate(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetHistory godoc
// @Summary     Get scan history for a QR code
// @Tags        scan-history
// @Produce     json
// @Param       id    path  string true "QR code ID"
// @Param       page  query int    false "Page"
// @Param       limit query int    false "Limit"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id}/history [get]
func (h *QRHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	result, err := h.scanSvc.GetHistory(c.Request.Context(), id, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetStats godoc
// @Summary     Get scan statistics for a QR code
// @Tags        scan-history
// @Produce     json
// @Param       id path string true "QR code ID"
// @Success     200 {object} response.APIResponse
// @Router      /api/v1/qr/{id}/stats [get]
func (h *QRHandler) GetStats(c *gin.Context) {
	id := c.Param("id")
	stats, err := h.scanSvc.GetStats(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
