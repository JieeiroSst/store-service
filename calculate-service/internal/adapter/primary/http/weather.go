package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/calculate-service/common"
	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

// GetForecast: GET /api/v1/weather/forecast?lat=10.8&lon=106.6
func (h *Handler) GetForecast(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "lat must be a number"})
		return
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "lon must be a number"})
		return
	}

	forecast, err := h.weather.GetForecast(c.Request.Context(), model.Coordinate{Lat: lat, Lon: lon})
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, forecast)
}

// GetTide: GET /api/v1/weather/tide?station=8518750
func (h *Handler) GetTide(c *gin.Context) {
	station := c.Query("station")

	tide, err := h.weather.GetTide(c.Request.Context(), station)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tide)
}

// GetRadar: GET /api/v1/weather/radar
func (h *Handler) GetRadar(c *gin.Context) {
	radar, err := h.weather.GetRadar(c.Request.Context())
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, radar)
}

// GetCurrent: GET /api/v1/weather/current?location=ho-chi-minh-city
func (h *Handler) GetCurrent(c *gin.Context) {
	location := c.Query("location")

	snapshot, err := h.weather.GetCurrentConditions(c.Request.Context(), location)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

// ListLocations: GET /api/v1/weather/locations
func (h *Handler) ListLocations(c *gin.Context) {
	c.JSON(http.StatusOK, h.weather.ListTrackedLocations(c.Request.Context()))
}

func respondErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrInvalidCoordinate):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, common.ErrLocationNotTracked):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, common.ErrUpstreamUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
