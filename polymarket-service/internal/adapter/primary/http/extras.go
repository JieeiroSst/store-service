package http

import (
	"net/http"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type redeemReferralRequest struct {
	RefCode string `json:"ref_code" binding:"required"`
}

func (h *Handler) GetReferral(c *gin.Context) {
	summary, err := h.referral.Summary(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) RedeemReferral(c *gin.Context) {
	var req redeemReferralRequest
	if !bind(c, &req) {
		return
	}
	ref, err := h.referral.Redeem(c.Request.Context(), c.Param("userId"), req.RefCode)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ref)
}

type marketRewardsRequest struct {
	DailyPool int64 `json:"daily_pool"`
	MaxSpread int64 `json:"max_spread"`
	MinSize   int64 `json:"min_size"`
}

// SetMarketRewards (admin) configures a market's liquidity rewards. A zero
// daily_pool turns them off.
func (h *Handler) SetMarketRewards(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req marketRewardsRequest
	if !bind(c, &req) {
		return
	}
	m, err := h.rewards.SetMarketRewards(c.Request.Context(), id, req.DailyPool, req.MaxSpread, req.MinSize)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(m))
}

func (h *Handler) GetMarketRewards(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	r, err := h.rewards.MarketRewards(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// GetUserRewards: ?limit=&cursor= -> {items, next_cursor, is_last_page, total_paid}
func (h *Handler) GetUserRewards(c *gin.Context) {
	page, total, err := h.rewards.UserRewards(c.Request.Context(), c.Param("userId"), queryInt(c, "limit"), c.Query("cursor"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": page.Items, "next_cursor": page.NextCursor, "is_last_page": page.IsLastPage, "total_paid": total.TotalPaid,
	})
}

type convertRequest struct {
	UserID    string  `json:"user_id"`
	MarketIDs []int64 `json:"market_ids" binding:"required,min=1"`
	Amount    int64   `json:"amount" binding:"required"`
}

// ConvertPositions swaps NO shares in the named markets of a neg-risk event
// for YES shares in all the others (plus cash when more than one is given).
func (h *Handler) ConvertPositions(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req convertRequest
	if !bind(c, &req) {
		return
	}
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	res, err := h.exchange.Convert(c.Request.Context(), port.ConvertInput{
		EventID: eventID, UserID: user, MarketIDs: req.MarketIDs, Amount: req.Amount,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	yes := res.YesMarkets
	if yes == nil {
		yes = []int64{}
	}
	c.JSON(http.StatusCreated, gin.H{"cash": res.Cash, "yes_markets": yes})
}

type winnerRequest struct {
	WinnerMarketID int64 `json:"winner_market_id" binding:"required"`
}

func (h *Handler) ProposeEvent(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req winnerRequest
	if !bind(c, &req) {
		return
	}
	event, err := h.resolution.ProposeEvent(c.Request.Context(), eventID, req.WinnerMarketID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventView(event))
}

func (h *Handler) ResolveEvent(c *gin.Context) {
	eventID, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req winnerRequest
	if !bind(c, &req) {
		return
	}
	event, err := h.resolution.ResolveEvent(c.Request.Context(), eventID, req.WinnerMarketID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventView(event))
}

func (h *Handler) GetExchange(c *gin.Context) {
	s, err := h.accounts.ExchangeSummary(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

type fundRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Amount int64  `json:"amount" binding:"required"`
}

// FundExchange (admin) tops up exchange revenue, the pot rewards are paid from,
// out of the named user's wallet.
func (h *Handler) FundExchange(c *gin.Context) {
	var req fundRequest
	if !bind(c, &req) {
		return
	}
	s, err := h.accounts.FundExchange(c.Request.Context(), req.UserID, req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, s)
}
