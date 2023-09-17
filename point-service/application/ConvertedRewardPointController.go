package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type ConvertedRewardPointContrller struct {
	convertedRewardPointService service.ConvertedRewardPointService
}

func InitConvertedRewardPointRouter(router *gin.Engine) {

	convertedRewardPointContrller := ConvertedRewardPointContrller{
		convertedRewardPointService: service.InitConvertedRewardPointServiceImpl(""),
	}

	router.GET("/", convertedRewardPointContrller.GetConvertedRewardPointHandler)
	router.GET("/", convertedRewardPointContrller.GetConvertedRewardPointByIdHandler)
	router.GET("/", convertedRewardPointContrller.CreateConvertedRewardPointHandler)
	router.GET("/", convertedRewardPointContrller.UpdateConvertedRewardPointHandler)
}

func (r *ConvertedRewardPointContrller) GetConvertedRewardPointHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) GetConvertedRewardPointByIdHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) CreateConvertedRewardPointHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) UpdateConvertedRewardPointHandler(c *gin.Context) {

}
