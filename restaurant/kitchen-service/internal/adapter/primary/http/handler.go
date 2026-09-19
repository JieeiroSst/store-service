package http

import "github.com/JIeeiroSst/kitchen-service/internal/domain/port"

type Handler struct {
	kitchen  port.KitchenUsecase
	food     port.FoodUsecase
	category port.CategoryUsecase
}

func NewHandler(kitchen port.KitchenUsecase, food port.FoodUsecase, category port.CategoryUsecase) *Handler {
	return &Handler{
		kitchen:  kitchen,
		food:     food,
		category: category,
	}
}
