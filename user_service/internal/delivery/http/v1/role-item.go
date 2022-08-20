package v1

import (
	"strconv"

	"github.com/JIeeiroSst/user-service/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initRoleItemRoutes(api *gin.RouterGroup) {
	roleItem := api.Group("/role-item")

	roleItem.PUT("/", h.UpdateRole)
	roleItem.DELETE("/", h.RemoveRole)
	roleItem.POST("/", h.AddRole)
}

func (h *Handler) AddRole(c *gin.Context) {
	var userRole model.UserRoles
	if err := c.ShouldBind(&userRole); err != nil {
		c.JSON(400, gin.H{})
		return
	}

	if err := h.usecase.UserRoles.AddRole(userRole); err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(400, gin.H{})
}

func (h *Handler) RemoveRole(c *gin.Context) {
	userId, err := strconv.Atoi(c.Query("user-id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	if err := h.usecase.UserRoles.RemoveRole(userId); err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(400, gin.H{})
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	userId, err := strconv.Atoi(c.Query("user-id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	roleId, err := strconv.Atoi(c.Query("role-id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	if err := h.usecase.UserRoles.Update(userId, roleId); err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(400, gin.H{})
}
