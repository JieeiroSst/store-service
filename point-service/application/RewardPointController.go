package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type RewardPointController struct {
	convertedRewardPointService service.RewardPointService
}

func InitRewardPointRouter(router *gin.Engine) {
	rewardPointController := RewardPointController{
		convertedRewardPointService: service.InitRewardPointServiceImpl(""),
	}

	router.GET("/", rewardPointController.GetRewardPointHandler)
	router.GET("/", rewardPointController.GetRewardPointByIDHandler)
	router.GET("/", rewardPointController.CreateRewardPointHandler)
	router.GET("/", rewardPointController.UpdateRewardPointHandler)
}

func (r *RewardPointController) GetRewardPointHandler(c *gin.Context) {

}

func (r *RewardPointController) GetRewardPointByIDHandler(c *gin.Context) {

}

func (r *RewardPointController) CreateRewardPointHandler(c *gin.Context) {

}

func (r *RewardPointController) UpdateRewardPointHandler(c *gin.Context) {

}
