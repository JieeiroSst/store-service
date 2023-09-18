package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardPointController struct {
	convertedRewardPointService service.RewardPointService
}

func InitRewardPointRouter(router *gin.Engine, dsn string) {
	rewardPointController := RewardPointController{
		convertedRewardPointService: service.InitRewardPointServiceImpl(dsn),
	}

	router.GET("/", rewardPointController.GetRewardPointHandler)
	router.GET("/:id", rewardPointController.GetRewardPointByIDHandler)
	router.POST("/", rewardPointController.CreateRewardPointHandler)
	router.PUT("/", rewardPointController.UpdateRewardPointHandler)
}

func (r *RewardPointController) GetRewardPointHandler(c *gin.Context) {
	// perPage := c.Query("per_page")
	// sortOrder := c.Query("sort_order")
	// cursor := c.Query("cursor")

	// response, err := r.
}

func (r *RewardPointController) GetRewardPointByIDHandler(c *gin.Context) {
	
}

func (r *RewardPointController) CreateRewardPointHandler(c *gin.Context) {

}

func (r *RewardPointController) UpdateRewardPointHandler(c *gin.Context) {

}
