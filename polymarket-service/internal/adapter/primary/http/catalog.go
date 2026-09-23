package http

import (
	"net/http"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createMarketRequest struct {
	Slug           string    `json:"slug"`
	Question       string    `json:"question" binding:"required"`
	GroupItemTitle string    `json:"group_item_title"`
	Description    string    `json:"description"`
	EndTime        time.Time `json:"end_time"`
	ShareValue     int64     `json:"share_value"`
	MinOrderSize   int64     `json:"min_order_size"`

	RewardPool      int64 `json:"reward_pool"`
	RewardMaxSpread int64 `json:"reward_max_spread"`
	RewardMinSize   int64 `json:"reward_min_size"`
}

type createEventRequest struct {
	Slug        string                `json:"slug"`
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	Category    string                `json:"category"`
	Tags        []string              `json:"tags"`
	ImageURL    string                `json:"image_url"`
	Featured    bool                  `json:"featured"`
	NegRisk     bool                  `json:"neg_risk"`
	EndDate     time.Time             `json:"end_date" binding:"required"`
	Markets     []createMarketRequest `json:"markets" binding:"required,min=1,dive"`
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var req createEventRequest
	if !bind(c, &req) {
		return
	}
	in := port.CreateEventInput{
		Slug: req.Slug, Title: req.Title, Description: req.Description, Category: req.Category, Tags: req.Tags,
		ImageURL: req.ImageURL, Featured: req.Featured, NegRisk: req.NegRisk, EndDate: req.EndDate,
	}
	for _, m := range req.Markets {
		in.Markets = append(in.Markets, port.CreateMarketInput{
			Slug: m.Slug, Question: m.Question, GroupItemTitle: m.GroupItemTitle, Description: m.Description,
			EndTime: m.EndTime, ShareValue: m.ShareValue, MinOrderSize: m.MinOrderSize,
			RewardPool: m.RewardPool, RewardMaxSpread: m.RewardMaxSpread, RewardMinSize: m.RewardMinSize,
		})
	}
	event, err := h.catalog.CreateEvent(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toEventView(event))
}

func (h *Handler) GetEvent(c *gin.Context) {
	event, err := h.catalog.GetEvent(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventView(event))
}

func (h *Handler) ListEvents(c *gin.Context) {
	page, err := h.catalog.ListEvents(c.Request.Context(), port.EventFilter{
		Status: model.EventStatus(c.Query("status")), Category: c.Query("category"), Tag: c.Query("tag"),
		Query: c.Query("q"), Featured: c.Query("featured") == "true", Sort: c.Query("sort"),
		Limit: queryInt(c, "limit"), Cursor: c.Query("cursor"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toEventViews(page.Items), "next_cursor": page.NextCursor, "is_last_page": page.IsLastPage})
}

func (h *Handler) RelatedEvents(c *gin.Context) {
	items, err := h.catalog.RelatedEvents(c.Request.Context(), c.Param("id"), queryInt(c, "limit"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventViews(items))
}

func (h *Handler) Categories(c *gin.Context) {
	items, err := h.catalog.Categories(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetMarket(c *gin.Context) {
	market, err := h.catalog.GetMarket(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(market))
}

func (h *Handler) SearchMarkets(c *gin.Context) {
	list, err := h.catalog.SearchMarkets(c.Request.Context(), port.MarketFilter{
		Query: c.Query("q"), Status: model.MarketStatus(c.Query("status")),
		Limit: queryInt(c, "limit"), Offset: queryInt(c, "offset"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]marketView, len(list.Items))
	for i := range list.Items {
		items[i] = toMarketView(&list.Items[i])
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": list.Total})
}
