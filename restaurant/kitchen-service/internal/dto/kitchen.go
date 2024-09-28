package dto

import (
	"strconv"

	"github.com/JIeeiroSst/kitchen-service/internal/model"
	"github.com/JieeiroSst/logger"
)

type Kitchen struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
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

func BuildCreateCategory(d Category) model.Category {
	return model.Category{
		ID:   logger.GearedIntID(),
		Name: d.Name,
	}
}

func BuildCategory(d Category) model.Category {
	return model.Category{
		ID:   d.ID,
		Name: d.Name,
	}
}

func BuildCreateFood(d Food) model.Food {
	return model.Food{
		ID:         logger.GearedIntID(),
		Name:       d.Name,
		CategoryID: d.CategoryID,
		Category:   BuildCategory(d.Category),
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

func BuildCreateKitchen(d Kitchen) model.Kitchen {
	return model.Kitchen{
		ID:    logger.GearedIntID(),
		Name:  d.Name,
		Foods: BuildFoods(d.Foods),
	}
}

func BuildKitchen(d Kitchen) model.Kitchen {
	return model.Kitchen{
		ID:    d.ID,
		Name:  d.Name,
		Foods: BuildFoods(d.Foods),
	}
}

func BuildModelFood(d model.Food) Food {
	return Food{
		ID:         d.ID,
		Name:       d.Name,
		CategoryID: d.CategoryID,
		Category:   BuildDtoCategory(d.Category),
	}
}

func BuildModelFoods(d []model.Food) []Food {
	var foods []Food
	for _, d := range d {
		foods = append(foods, BuildModelFood(d))
	}
	return foods
}

func BuildDtoCategory(d model.Category) Category {
	return Category{
		ID:   d.ID,
		Name: d.Name,
	}
}

func BuildDtoCategories(d []model.Category) []Category {
	categories := make([]Category, 0)
	for _, v := range d {
		categories = append(categories, BuildDtoCategory(v))
	}
	return categories
}

type Order struct {
	TableName string   `json:"table_name" form:"table_name"`
	KitchenID int      `json:"kitchen_id"`
	MenuIDs   []string `json:"menu_id" form:"menu_id"`
}

func (k Kitchen) Build() Order {
	menuIds := make([]string, 0)
	if len(k.Foods) > 0 {
		for _, v := range k.Foods {
			menuIds = append(menuIds, strconv.Itoa(v.ID))
		}
	}
	return Order{
		TableName: k.Name,
		KitchenID: k.ID,
		MenuIDs:   menuIds,
	}
}

type Customer struct {
	Name      string     `json:"table_name" form:"table_name"`
	KitchenID int        `json:"kitchen_id"`
	Menu      []MenuFood `json:"menu" form:"menu"`
}

type MenuFood struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Category Category `json:"category"`
}

func (k Customer) Build() Kitchen {
	foods := make([]Food, 0)
	for _, v := range k.Menu {
		foods = append(foods, Food{
			Name: v.Name,
			Category: Category{
				ID:   v.Category.ID,
				Name: v.Category.Name,
			},
		})
	}
	return Kitchen{
		Name:  k.Name,
		Foods: foods,
	}
}
