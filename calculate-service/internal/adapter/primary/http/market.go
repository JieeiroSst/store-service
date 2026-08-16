package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMarketSnapshot: GET /api/v1/market/snapshot
// Returns the latest polled CoinGecko market data for the electronic board.
func (h *Handler) GetMarketSnapshot(c *gin.Context) {
	snapshot, err := h.market.GetSnapshot(c.Request.Context())
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

// MarketWS: GET /api/v1/market/ws
// Upgrades to a WebSocket and streams a fresh snapshot on every poll tick.
func (h *Handler) MarketWS(c *gin.Context) {
	h.hub.ServeWS(c)
}
