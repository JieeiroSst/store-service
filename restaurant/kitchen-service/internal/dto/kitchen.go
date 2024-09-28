package dto

import "github.com/JIeeiroSst/kitchen-service/internal/model"

type Kitchen struct {
	ID    int    `json:"id"`
	Foods []Food `json:"foods"`
}

type Food struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Category   Category `json:"category"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func BuildCategory(d Category) model.Category {
	return model.Category{
		ID:   d.ID,
		Name: d.Name,
	}
}

func BuildFood(d Food) model.Food {
	return model.Food{
		ID:         d.ID,
		Name:       d.Name,
		CategoryID: d.CategoryID,
		Category:   BuildCategory(d.Category),
	}
}

func BuildFoods(d []Food) []model.Food {
	var foods []model.Food
	for _, d := range d {
		foods = append(foods, BuildFood(d))
	}
	return foods
}

func BuildKitchen(d Kitchen) model.Kitchen {
	return model.Kitchen{
		ID:    d.ID,
		Foods: BuildFoods(d.Foods),
	}
}
