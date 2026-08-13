package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/gin-gonic/gin"
)

type QRHandler struct {
	qr              *app.QRService
	shortlinkDomain string
}

func NewQRHandler(qr *app.QRService, shortlinkDomain string) *QRHandler {
	return &QRHandler{qr, shortlinkDomain}
}

func (h *QRHandler) Generate(c *gin.Context) {
	format := c.DefaultQuery("format", "png")
	size, _ := strconv.Atoi(c.DefaultQuery("size", "512"))
	color := c.DefaultQuery("color", "#000000")
	bgcolor := c.DefaultQuery("bgcolor", "#ffffff")

	shortDomain := h.shortlinkDomain
	if shortDomain == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		shortDomain = scheme + "://" + c.Request.Host
	}

	result, err := h.qr.Generate(c.Request.Context(), app.QRInput{
		LinkID: c.Param("id"), Format: format, Size: size, Color: color, BGColor: bgcolor,
		ShortLinkDomain: shortDomain,
	})
	if err != nil {
		switch {
		case errors.Is(err, app.ErrInvalidQRFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": `Invalid format. Use "png" or "svg".`})
		case errors.Is(err, app.ErrLinkNotFound):
			respondNotFound(c, "Link not found")
		default:
			respondInternalError(c, "Failed to generate QR code", err)
		}
		return
	}

	filename := "qr-code"
	if result.ShortCode != "" {
		filename = "qr-" + result.ShortCode
	}

	if result.ContentType == "image/png" {
		c.Header("Cache-Control", "public, max-age=86400")
		c.Header("Content-Disposition", `inline; filename="`+filename+`.png"`)
		c.Data(http.StatusOK, "image/png", result.PNG)
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Disposition", `inline; filename="`+filename+`.svg"`)
	c.Data(http.StatusOK, "image/svg+xml", []byte(result.SVG))
}
