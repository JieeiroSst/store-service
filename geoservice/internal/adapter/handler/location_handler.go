package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/geoservice/internal/core/domain"
	"github.com/geoservice/internal/core/port"
	"github.com/geoservice/internal/core/service"
)

const defaultToleranceM = 1000.0

type LocationHandler struct {
	svc port.LocatorService
}

func NewLocationHandler(svc port.LocatorService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

func (h *LocationHandler) Register(e *echo.Echo) {
	g := e.Group("/api/v1")
	g.GET("/locate", h.Locate)
	g.GET("/provinces", h.ListProvinces)
	g.GET("/provinces/:code", h.GetProvince)
	g.GET("/provinces/:code/districts", h.ListDistricts)
	g.GET("/districts/:code/wards", h.ListWards)
	g.GET("/provinces/:code/streets", h.ListStreets)
	g.GET("/stats", h.Stats)
	g.GET("/whitelist/check", h.CheckWhitelist)
	g.GET("/whitelist/cities", h.ListWhitelistCities)
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}

// Locate: GET /api/v1/locate?lng=106.70&lat=10.77&tolerance=500
// Returns the full province → district → ward chain containing the point.
func (h *LocationHandler) Locate(c echo.Context) error {
	lng, err := strconv.ParseFloat(c.QueryParam("lng"), 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "lng must be a number")
	}
	lat, err := strconv.ParseFloat(c.QueryParam("lat"), 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "lat must be a number")
	}
	tolerance := defaultToleranceM
	if t := c.QueryParam("tolerance"); t != "" {
		tolerance, _ = strconv.ParseFloat(t, 64)
	}

	res, err := h.svc.Locate(c.Request().Context(), domain.Point{Lng: lng, Lat: lat}, tolerance)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCoordinate) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !res.Found {
		return c.JSON(http.StatusNotFound, res)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *LocationHandler) ListProvinces(c echo.Context) error {
	list, err := h.svc.ListProvinces(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, list)
}

func (h *LocationHandler) GetProvince(c echo.Context) error {
	prov, err := h.svc.GetProvince(c.Request().Context(), c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if prov == nil {
		return echo.NewHTTPError(http.StatusNotFound, "province not found")
	}
	return c.JSON(http.StatusOK, prov)
}

func (h *LocationHandler) ListDistricts(c echo.Context) error {
	list, err := h.svc.ListDistricts(c.Request().Context(), c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, list)
}

func (h *LocationHandler) ListWards(c echo.Context) error {
	list, err := h.svc.ListWards(c.Request().Context(), c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, list)
}

func (h *LocationHandler) ListStreets(c echo.Context) error {
	list, err := h.svc.ListStreets(c.Request().Context(), c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, list)
}

// CheckWhitelist: GET /api/v1/whitelist/check?lng=106.70&lat=10.77&tolerance=1000
// Trả về whitelisted=true nếu tọa độ nằm trong thành phố được ưu tiên,
// kèm thông tin tỉnh/thành phố, quận/huyện, phường/xã, đường.
func (h *LocationHandler) CheckWhitelist(c echo.Context) error {
	lng, err := strconv.ParseFloat(c.QueryParam("lng"), 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "lng must be a number")
	}
	lat, err := strconv.ParseFloat(c.QueryParam("lat"), 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "lat must be a number")
	}
	tolerance := defaultToleranceM
	if t := c.QueryParam("tolerance"); t != "" {
		tolerance, _ = strconv.ParseFloat(t, 64)
	}

	res, err := h.svc.CheckWhitelist(c.Request().Context(), domain.Point{Lng: lng, Lat: lat}, tolerance)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCoordinate) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}

// ListWhitelistCities: GET /api/v1/whitelist/cities
// Trả về danh sách các thành phố trong whitelist.
func (h *LocationHandler) ListWhitelistCities(c echo.Context) error {
	list, err := h.svc.ListWhitelistCities(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, list)
}

func (h *LocationHandler) Stats(c echo.Context) error {
	p, d, w, err := h.svc.CountSeeded(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]int64{
		"provinces": p, "districts": d, "wards": w,
	})
}
