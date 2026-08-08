package http

import "github.com/JIeeiroSst/calculate-service/internal/domain/port"

type Handler struct {
	weather port.WeatherService
}

func NewHandler(weather port.WeatherService) *Handler {
	return &Handler{weather: weather}
}
