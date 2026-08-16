package http

import (
	"github.com/JIeeiroSst/calculate-service/internal/adapter/primary/ws"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
)

type Handler struct {
	weather port.WeatherService
	market  port.MarketService
	hub     *ws.Hub
}

func NewHandler(weather port.WeatherService, market port.MarketService, hub *ws.Hub) *Handler {
	return &Handler{weather: weather, market: market, hub: hub}
}
