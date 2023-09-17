package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardDiscountController struct {
	rewardDiscountService service.RewardDiscountService
}

func InitRewardDiscountRouter(router *gin.Engine, dsn string) {
	rewardDiscountController := RewardDiscountController{
		rewardDiscountService: service.InitRewardDiscountServiceImpl(dsn),
	}
	router.GET("/", rewardDiscountController.GetRewardDiscountHandler)
	router.GET("/:id", rewardDiscountController.GetRewardDiscountByIdHandler)
	router.POST("/", rewardDiscountController.CreateRewardDiscountHandler)
	router.PUT("/", rewardDiscountController.UpdateRewardDiscountHandler)
}

func (r *RewardDiscountController) GetRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardDiscountController) GetRewardDiscountByIdHandler(c *gin.Context) {

}

func (r *RewardDiscountController) CreateRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardDiscountController) UpdateRewardDiscountHandler(c *gin.Context) {

}
