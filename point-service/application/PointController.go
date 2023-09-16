package application

import (
	"github.com/gin-gonic/gin"
)

type PointController struct {
}

func InitRestaurantController(router *gin.Engine) {
	restaurantController := PointController{
		// restaurantService: service.InitRestaurantServiceImpl(),
	}
	router.GET("/food", restaurantController.GetPointHandler)
}

func (r *PointController) GetPointHandler(c *gin.Context) {
	// var response dto.Response

	// foods, response := r.restaurantService.GetFoods()

	// if response.Status != http.StatusOK {
	// 	c.JSON(response.Status, response)
	// 	return
	// }
	// c.JSON(response.Status, foods)
}
