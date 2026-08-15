package http

import (
	stdhttp "net/http"

	fileapp "github.com/JIeeiroSst/voucher-service/internal/application/file"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	svc fileapp.FileService
}

func NewFileHandler(svc fileapp.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

func (h *FileHandler) ImportStock(c *gin.Context) {
	merchantID, err := shared.ParseMerchantID(c.PostForm("merchant_id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid merchant_id"}})
		return
	}
	productSKU := c.PostForm("product_sku")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "file is required"}})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "could not open file"}})
		return
	}
	defer f.Close()

	result, err := h.svc.ImportStockCSV(c.Request.Context(), fileapp.ImportStockInput{
		MerchantID: merchantID, ProductSKU: productSKU, Reader: f,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, result)
}
