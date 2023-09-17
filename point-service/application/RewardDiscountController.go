package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardDiscountController struct {
	rewardDiscountService service.RewardDiscountService
}

func InitRewardDiscountRouter(router *gin.Engine) {
	rewardDiscountController := RewardDiscountController{
		rewardDiscountService: service.InitRewardDiscountServiceImpl(""),
	}
	router.GET("/", rewardDiscountController.GetRewardDiscountHandler)
	router.GET("/", rewardDiscountController.GetRewardDiscountByIdHandler)
	router.GET("/", rewardDiscountController.CreateRewardDiscountHandler)
	router.GET("/", rewardDiscountController.UpdateRewardDiscountHandler)
}

func (r *RewardDiscountController) GetRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardDiscountController) GetRewardDiscountByIdHandler(c *gin.Context) {

}

func (r *RewardDiscountController) CreateRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardDiscountController) UpdateRewardDiscountHandler(c *gin.Context) {

}
