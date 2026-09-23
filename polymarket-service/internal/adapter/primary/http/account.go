package http

import (
	"net/http"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type amountRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

func (h *Handler) Deposit(c *gin.Context) {
	var req amountRequest
	if !bind(c, &req) {
		return
	}
	b, err := h.accounts.Deposit(c.Request.Context(), c.Param("userId"), req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var req amountRequest
	if !bind(c, &req) {
		return
	}
	b, err := h.accounts.Withdraw(c.Request.Context(), c.Param("userId"), req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetBalance(c *gin.Context) {
	b, err := h.accounts.Balance(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) GetWallet(c *gin.Context) {
	w, err := h.accounts.Wallet(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallet_id": w.WalletID, "balance": w.Balance, "currency": w.Currency, "status": w.Status})
}

func (h *Handler) GetPortfolio(c *gin.Context) {
	p, err := h.accounts.Portfolio(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) ClosedPositions(c *gin.Context) {
	items, err := h.accounts.ClosedPositions(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetActivity(c *gin.Context) {
	items, err := h.accounts.Activity(c.Request.Context(), c.Param("userId"), queryInt(c, "limit"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// ListUserOrders: ?market_id=&status=open|filled|canceled|expired&limit=&cursor= -> {items, next_cursor, is_last_page}
func (h *Handler) ListUserOrders(c *gin.Context) {
	orders, err := h.exchange.ListOrders(c.Request.Context(), port.OrderFilter{
		UserID: c.Param("userId"), MarketID: int64(queryInt(c, "market_id")), Status: model.OrderStatus(c.Query("status")),
		Limit: queryInt(c, "limit"), Cursor: c.Query("cursor"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}
