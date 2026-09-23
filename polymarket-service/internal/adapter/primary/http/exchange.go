package http

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type placeOrderRequest struct {
	UserID        string     `json:"user_id"`
	ClientOrderID string     `json:"client_order_id"`
	Outcome       string     `json:"outcome" binding:"required"`
	Side          string     `json:"side" binding:"required"`
	Type          string     `json:"type"`
	TimeInForce   string     `json:"time_in_force"`
	Price         int64      `json:"price"`
	Size          int64      `json:"size"`
	Amount        int64      `json:"amount"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

func (h *Handler) PlaceOrder(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req placeOrderRequest
	if !bind(c, &req) {
		return
	}
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	res, err := h.exchange.PlaceOrder(c.Request.Context(), port.PlaceOrderInput{
		MarketID: marketID, UserID: user, ClientOrderID: req.ClientOrderID,
		Outcome: model.Outcome(req.Outcome), Side: model.Side(req.Side),
		Type: model.OrderType(req.Type), TimeInForce: model.TimeInForce(req.TimeInForce),
		Price: req.Price, Size: req.Size, Amount: req.Amount, ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	trades := res.Trades
	if trades == nil {
		trades = []model.Trade{}
	}
	c.JSON(http.StatusCreated, gin.H{"order": res.Order, "trades": trades})
}

func (h *Handler) GetOrder(c *gin.Context) {
	id, ok := pathInt(c, "orderId")
	if !ok {
		return
	}
	order, err := h.exchange.GetOrder(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id, ok := pathInt(c, "orderId")
	if !ok {
		return
	}
	user, ok := h.caller(c, c.Query("user_id"))
	if !ok {
		return
	}
	order, err := h.exchange.CancelOrder(c.Request.Context(), id, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) CancelAllOrders(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	user, ok := h.caller(c, c.Query("user_id"))
	if !ok {
		return
	}
	n, err := h.exchange.CancelAll(c.Request.Context(), marketID, user)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"canceled": n})
}

func (h *Handler) GetBook(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	book, err := h.exchange.OrderBook(c.Request.Context(), marketID, outcomeParam(c, model.OutcomeYes))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, book)
}

func (h *Handler) GetQuote(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	amount := int64(queryInt(c, "amount"))
	q, err := h.exchange.Quote(c.Request.Context(), port.QuoteInput{
		MarketID: marketID, Outcome: outcomeParam(c, model.OutcomeYes), Side: model.Side(c.Query("side")), Amount: amount,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shares": q.Shares, "cash": q.Cash, "fee": q.Fee, "avg_price": q.AvgPrice, "worst_price": q.WorstPrice, "fillable": q.Fillable,
	})
}

func (h *Handler) ListMarketTrades(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	trades, err := h.exchange.MarketTrades(c.Request.Context(), marketID, queryInt(c, "limit"), queryInt(c, "offset"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, trades)
}

// GetPriceHistory: ?outcome=&range=1h|6h|1d|1w|1m|max&bucket=60 (seconds)
func (h *Handler) GetPriceHistory(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	candles, err := h.exchange.PriceHistory(c.Request.Context(), port.PriceHistoryInput{
		MarketID: marketID, Outcome: outcomeParam(c, model.OutcomeYes), Range: c.Query("range"),
		Bucket: time.Duration(queryInt(c, "bucket")) * time.Second,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, candles)
}

func (h *Handler) TopHolders(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	holders, err := h.social.TopHolders(c.Request.Context(), marketID, outcomeParam(c, model.OutcomeYes), queryInt(c, "limit"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, holders)
}

// Stream pushes "book", "trade", "status" and "resolved" events for a market as
func (h *Handler) Stream(c *gin.Context) {
	marketID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	snapshot, err := h.exchange.OrderBook(c.Request.Context(), marketID, model.OutcomeYes)
	if err != nil {
		writeError(c, err)
		return
	}
	events, cancel := h.stream.Subscribe(port.MarketTopic(marketID))
	defer cancel()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	c.SSEvent("book", snapshot)
	c.Writer.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	ctx := c.Request.Context()
	c.Stream(func(io.Writer) bool { return h.pump(ctx, c, events, heartbeat) })
}

func (h *Handler) pump(ctx context.Context, c *gin.Context, events <-chan port.StreamMessage, heartbeat *time.Ticker) bool {
	select {
	case <-ctx.Done():
		return false
	case msg, ok := <-events:
		if !ok {
			return false
		}
		c.SSEvent(msg.Type, msg.Payload)
	case <-heartbeat.C:
		c.SSEvent("ping", "")
	}
	return true
}
