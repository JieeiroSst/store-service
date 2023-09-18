package application

import (
	"github.com/JIeeiroSst/point-service/domain/dto"
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
	perPage := c.Query("per_page")
	sortOrder := c.Query("sort_order")
	cursor := c.Query("cursor")

	response, err := r.rewardDiscountService.GetAll(c, perPage, sortOrder, cursor)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, response)
}

func (r *RewardDiscountController) GetRewardDiscountByIdHandler(c *gin.Context) {
	id := c.Param("id")

	response, err := r.rewardDiscountService.GetByID(c, id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, response)
}

func (r *RewardDiscountController) CreateRewardDiscountHandler(c *gin.Context) {
	var data dto.RewardDiscountDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err)
		return
	}

	if err := r.rewardDiscountService.Create(c, data); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "create success")
}

func (r *RewardDiscountController) UpdateRewardDiscountHandler(c *gin.Context) {
	var data dto.RewardDiscountDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, err)
		return
	}

	if err := r.rewardDiscountService.Update(c, data); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "update success")
}
