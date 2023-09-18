package application

import (
	"github.com/JIeeiroSst/point-service/domain/dto"
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
	perPage := c.Query("per_page")
	sortOrder := c.Query("sort_order")
	cursor := c.Query("cursor")

	resp, err := r.convertedRewardPointService.GetAll(c, perPage, sortOrder, cursor)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, resp)
}

func (r *RewardPointController) GetRewardPointByIDHandler(c *gin.Context) {
	id := c.Param("id")
	resp, err := r.convertedRewardPointService.GetByID(c, id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, resp)
}

func (r *RewardPointController) CreateRewardPointHandler(c *gin.Context) {
	var data dto.RewardPointDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err)
		return
	}

	if err := r.convertedRewardPointService.Create(c, data); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "create success")
}

func (r *RewardPointController) UpdateRewardPointHandler(c *gin.Context) {
	var data dto.RewardPointDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err)
		return
	}

	if err := r.convertedRewardPointService.Update(c, data); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "update success")
}
