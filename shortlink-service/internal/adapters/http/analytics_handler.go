package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analytics *app.AnalyticsService
}

func NewAnalyticsHandler(analytics *app.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics}
}

func queryDays(c *gin.Context) int {
	days := 30
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	return days
}

func dateCountsJSON(rows []ports.DateCount) []gin.H {
	out := make([]gin.H, len(rows))
	for i, r := range rows {
		out[i] = gin.H{"date": r.Date, "clicks": r.Clicks}
	}
	return out
}

func countryCountsJSON(rows []ports.CountryCount) []gin.H {
	out := make([]gin.H, len(rows))
	for i, r := range rows {
		out[i] = gin.H{"countryCode": r.CountryCode, "country": r.Country, "clicks": r.Clicks}
	}
	return out
}

func deviceCountsJSON(rows []ports.DeviceCount) []gin.H {
	out := make([]gin.H, len(rows))
	for i, r := range rows {
		out[i] = gin.H{"device": r.Device, "clicks": r.Clicks}
	}
	return out
}

func platformCountsJSON(rows []ports.PlatformCount) []gin.H {
	out := make([]gin.H, len(rows))
	for i, r := range rows {
		out[i] = gin.H{"platform": r.Platform, "clicks": r.Clicks}
	}
	return out
}

func (h *AnalyticsHandler) Overview(c *gin.Context) {
	result, err := h.analytics.Overview(c.Request.Context(), optionalUserID(c), queryDays(c))
	if err != nil {
		respondInternalError(c, "Failed to get analytics overview", err)
		return
	}

	topLinks := make([]gin.H, len(result.TopLinks))
	for i, l := range result.TopLinks {
		topLinks[i] = gin.H{
			"id": l.ID, "shortCode": l.ShortCode, "title": l.Title, "originalUrl": l.OriginalURL,
			"totalClicks": l.TotalClicks, "uniqueClicks": l.UniqueClicks,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"totalClicks": result.TotalClicks, "uniqueClicks": result.UniqueClicks,
		"clicksByDate": dateCountsJSON(result.ClicksByDate), "clicksByCountry": countryCountsJSON(result.ClicksByCountry),
		"clicksByDevice": deviceCountsJSON(result.ClicksByDevice), "clicksByPlatform": platformCountsJSON(result.ClicksByPlatform),
		"topLinks": topLinks,
	})
}

func (h *AnalyticsHandler) ForLink(c *gin.Context) {
	result, err := h.analytics.ForLink(c.Request.Context(), c.Param("linkId"), optionalUserID(c), queryDays(c))
	if err != nil {
		if errors.Is(err, app.ErrLinkNotFound) {
			respondNotFound(c, "Link not found")
			return
		}
		respondInternalError(c, "Failed to get link analytics", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalClicks": result.TotalClicks, "uniqueClicks": result.UniqueClicks,
		"clicksByDate": dateCountsJSON(result.ClicksByDate), "clicksByCountry": countryCountsJSON(result.ClicksByCountry),
		"clicksByDevice": deviceCountsJSON(result.ClicksByDevice), "clicksByPlatform": platformCountsJSON(result.ClicksByPlatform),
	})
}
