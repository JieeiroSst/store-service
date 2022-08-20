package v1

import (
	"strconv"

	"github.com/JIeeiroSst/user-service/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initRoleRoutes(api *gin.RouterGroup) {
	roleGroup := api.Group("/role")
	roleGroup.GET("/:id", h.Role)
	roleGroup.GET("/", h.Roles)
	roleGroup.POST("/", h.CreateRole)
	roleGroup.PUT("/", h.UpdateRole)
	roleGroup.DELETE("/", h.DeleteRole)
}

func (h *Handler) CreateRole(c *gin.Context) {
	name := c.Query("name")
	role := model.Role{
		Name: name,
	}
	if err := h.usecase.Roles.Create(role); err != nil {
		c.JSON(500, gin.H{})
		return
	}
	c.JSON(200, gin.H{})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	name := c.Query("name")

	if err := h.usecase.Roles.Update(id, name); err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(200, gin.H{})
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	if err := h.usecase.Roles.Delete(id); err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(200, gin.H{})
}

func (h *Handler) Role(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	role, err := h.usecase.Roles.Role(id)
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(200, role)
}

func (h *Handler) Roles(c *gin.Context) {
	roles, err := h.usecase.Roles.Roles()
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(200, roles)
}
