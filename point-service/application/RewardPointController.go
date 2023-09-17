package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardPointController struct {
	convertedRewardPointService service.ConvertedRewardPointService
}

func InitRewardPointController(convertedRewardPointService service.ConvertedRewardPointService) *RewardPointController {
	return &RewardPointController{
		convertedRewardPointService: convertedRewardPointService,
	}
}

func InitRewardPointRouter(router *gin.Engine) {
}

func (r *RewardPointController) GetRewardPointHandler(c *gin.Context) {

}
