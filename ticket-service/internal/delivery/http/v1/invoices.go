package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initInvoices(api *gin.RouterGroup) {
	_ = api.Group("/invoices")

	
}
