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

// CreateRole godoc
// @Summary Create Role
// @Description Create Role
// @Accept  json
// @Produce  json
// @Param name  query string true "name in role"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role [post]
func (h *Handler) CreateRole(c *gin.Context) {
	name := c.Query("name")
	role := model.Role{
		Name: name,
	}
	if err := h.usecase.Roles.Create(role); err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "create role success",
	})
}

// UpdateRole godoc
// @Summary Update Role
// @Description Update Role
// @Accept  json
// @Produce  json
// @Param id  query string true "id role"
// @Param name  query string true "name in role"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, gin.H{})
		return
	}
	name := c.Query("name")

	if err := h.usecase.Roles.Update(id, name); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "update role success",
	})
}

// DeleteRole godoc
// @Summary Delete Role
// @Description Delete Role
// @Accept  json
// @Produce  json
// @Param id  query string true "id role"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err := h.usecase.Roles.Delete(id); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "delete role success",
	})
}

// Role godoc
// @Summary Role
// @Description Role
// @Accept  json
// @Produce  json
// @Param id  path string true "id role"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role/:id [get]
func (h *Handler) Role(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	role, err := h.usecase.Roles.Role(id)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, role)
}

// Roles godoc
// @Summary Roles
// @Description Roles
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role [get]
func (h *Handler) Roles(c *gin.Context) {
	roles, err := h.usecase.Roles.Roles()
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, roles)
}
