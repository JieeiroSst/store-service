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

// AddRole godoc
// @Summary AddRole
// @Description AddRole
// @Accept  json
// @Produce  json
// @Param user-role  body model.UserRoles true "user role in json"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role-item [post]
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

// Login godoc
// @Summary Login Account
// @Description login account
// @Accept  json
// @Produce  json
// @Param user-id  query string true "user-id"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role-item [delete]
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

// UpdateUserRole godoc
// @Summary UpdateUser Role
// @Description UpdateUser Role
// @Accept  json
// @Produce  json
// @Param user-id  query string true "user-id"
// @Param role-id  query string true "role-id"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/role-item [put]
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
