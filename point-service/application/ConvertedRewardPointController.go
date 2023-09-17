package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type ConvertedRewardPointContrller struct {
	convertedRewardPointService service.ConvertedRewardPointService
}

func InitConvertedRewardPointContrller(convertedRewardPointService service.ConvertedRewardPointService) *ConvertedRewardPointContrller {
	return &ConvertedRewardPointContrller{
		convertedRewardPointService: convertedRewardPointService,
	}
}

func InitConvertedRewardPointRouter(router *gin.Engine) {

}

func (r *RewardPointController) GetConvertedRewardPointHandler(c *gin.Context) {

}

func (r *RewardPointController) GetConvertedRewardPointByIdHandler(c *gin.Context) {

}

func (r *RewardPointController) CreateConvertedRewardPointHandler(c *gin.Context) {

}

func (r *RewardPointController) UpdateConvertedRewardPointHandler(c *gin.Context) {

}
