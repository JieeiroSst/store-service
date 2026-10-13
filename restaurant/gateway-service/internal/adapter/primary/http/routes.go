package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.Default()
	engine.Use(CORSMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(nethttp.StatusOK, gin.H{"status": "ok"})
	})

	RegisterRoutes(engine, h)

	return engine
}

// RegisterRoutes forwards each prefix to its downstream service, stripping
// the gateway-level prefix so the backend sees its own native paths
// (e.g. "/consumer/api/v1/consumer" -> "/api/v1/consumer").
func RegisterRoutes(engine *gin.Engine, h *Handler) {
	api := engine.Group("", Authenticate(h.auth))
	api.Any("/consumer/*proxyPath", gin.WrapH(nethttp.StripPrefix("/consumer", h.gateway.ConsumerHandler())))
	api.Any("/accounting/*proxyPath", gin.WrapH(nethttp.StripPrefix("/accounting", h.gateway.AccountingHandler())))
	api.Any("/delivery/*proxyPath", gin.WrapH(nethttp.StripPrefix("/delivery", h.gateway.DeliveryHandler())))
}
