package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardDiscountController struct {
	rewardDiscountService service.RewardDiscountService
}

func InitRewardDiscountController(rewardDiscountService service.RewardDiscountService) *RewardDiscountController {
	return &RewardDiscountController{
		rewardDiscountService: rewardDiscountService,
	}
}

func InitRewardDiscountRouter(router *gin.Engine) {

}

func (r *RewardPointController) GetRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardPointController) GetRewardDiscountByIdHandler(c *gin.Context) {

}

func (r *RewardPointController) CreateRewardDiscountHandler(c *gin.Context) {

}

func (r *RewardPointController) UpdateRewardDiscountHandler(c *gin.Context) {

}
