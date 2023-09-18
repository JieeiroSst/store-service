package application

import (
	"github.com/JIeeiroSst/point-service/domain/dto"
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
	perPage := c.Query("per_page")
	sortOrder := c.Query("sort_order")
	cursor := c.Query("cursor")

	response, err := r.convertedRewardPointService.GetAll(c, perPage, sortOrder, cursor)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, response)
}

func (r *ConvertedRewardPointContrller) GetConvertedRewardPointByIdHandler(c *gin.Context) {
	id := c.Param("id")
	response, err := r.convertedRewardPointService.GetByID(c, id)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, response)
}

func (r *ConvertedRewardPointContrller) CreateConvertedRewardPointHandler(c *gin.Context) {
	var data dto.ConvertedRewardPointDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err.Error())
	}

	if err := r.convertedRewardPointService.Create(c, data); err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, "create success")
}

func (r *ConvertedRewardPointContrller) UpdateConvertedRewardPointHandler(c *gin.Context) {
	var data dto.ConvertedRewardPointDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err.Error())
	}

	if err := r.convertedRewardPointService.Update(c, data); err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, "update success")
}
