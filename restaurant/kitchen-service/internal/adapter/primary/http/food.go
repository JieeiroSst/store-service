package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FindFood(c *gin.Context) {
	var pagination logger.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	foods, err := h.food.Find(c, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": foods})
}

func (h *Handler) CreateFood(c *gin.Context) {
	var food model.Food
	if err := c.ShouldBind(&food); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.food.Create(c, &food); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// FindFoodByIDs is the price/name lookup order-service calls to compute a
// trusted order total from a set of food IDs, e.g. "?ids=1,2,3".
func (h *Handler) FindFoodByIDs(c *gin.Context) {
	raw := strings.Split(c.Query("ids"), ",")
	ids := make([]int, 0, len(raw))
	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id in ids: " + v})
			return
		}
		ids = append(ids, id)
	}

	foods, err := h.food.FindByIDs(c, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": foods})
}
