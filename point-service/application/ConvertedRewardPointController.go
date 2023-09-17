package application

import (
	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type ConvertedRewardPointContrller struct {
	convertedRewardPointService service.ConvertedRewardPointService
}

func InitConvertedRewardPointRouter(router *gin.Engine, dsn string) {

	convertedRewardPointContrller := ConvertedRewardPointContrller{
		convertedRewardPointService: service.InitConvertedRewardPointServiceImpl(dsn),
	}

	router.GET("/", convertedRewardPointContrller.GetConvertedRewardPointHandler)
	router.GET("/:id", convertedRewardPointContrller.GetConvertedRewardPointByIdHandler)
	router.POST("/", convertedRewardPointContrller.CreateConvertedRewardPointHandler)
	router.PUT("/", convertedRewardPointContrller.UpdateConvertedRewardPointHandler)
}

func (r *ConvertedRewardPointContrller) GetConvertedRewardPointHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) GetConvertedRewardPointByIdHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) CreateConvertedRewardPointHandler(c *gin.Context) {

}

func (r *ConvertedRewardPointContrller) UpdateConvertedRewardPointHandler(c *gin.Context) {

}
