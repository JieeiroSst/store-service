package v1

import (
	"errors"
	"strconv"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initCasbinRoutes(api *gin.RouterGroup) {
	group := api.Group("/casbin")

	group.POST("/authorize", h.Authorize)
	group.GET("/", h.CasbinRuleAll)
	group.GET("/:id", h.CasbinRuleById)
	group.POST("/", h.CreateCasbinRule)
	group.DELETE("/:id", h.DeleteCasbinRule)
	group.PUT("/:id/ptype/:ptype", h.UpdateCasbinRulePtype)
	group.PUT("/:id/name/:name", h.UpdateCasbinRuleName)
	group.PUT("/:id/endpoint/:endpoint", h.UpdateCasbinRuleEndpoint)
	group.PUT("/:id/method/:method", h.UpdateCasbinMethod)
}

// Authorize godoc
// @Summary Authorize Account
// @Description Authorize account
// @Accept  json
// @Produce  json
// @Param username query string true "username"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/authentication [post]
func (h *Handler) Authorize(ctx *gin.Context) {
	username := ctx.Query("username")
	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	auth := model.CasbinAuth{
		Sub: username,
		Obj: path,
		Act: method,
	}

	isOk, err := h.usecase.Casbins.EnforceCasbin(auth)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}

	if !isOk {
		ctx.JSON(401, gin.H{
			"code":    401,
			"message": "THE CUSTOMER IS NOT AUTHORIZED FOR THE CONTENT REQUESTED",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "THE CUSTOMER IS AUTHORIZED FOR THE CONTENT REQUESTED",
	})
}

// CasbinRuleAll godoc
// @Summary CasbinRuleAll Account
// @Description CasbinRuleAll account
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin [get]
func (h *Handler) CasbinRuleAll(ctx *gin.Context) {
	casbins, err := h.usecase.Casbins.CasbinRuleAll()
	if errors.Is(err, common.NotFound) {
		ctx.JSON(404, gin.H{
			"code":    404,
			"message": "NOT FOUND",
		})
		return
	}
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, casbins)
}

// CasbinRuleById godoc
// @Summary CasbinRuleById Account
// @Description CasbinRuleById account
// @Accept  json
// @Produce  json
// @Param id query int true "id"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id [get]
func (h *Handler) CasbinRuleById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	casbin, err := h.usecase.Casbins.CasbinRuleById(id)
	if errors.Is(err, common.NotFound) {
		ctx.JSON(404, gin.H{
			"code":    404,
			"message": "NOT FOUND",
		})
		return
	}
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, casbin)
}

// CreateCasbinRule godoc
// @Summary CreateCasbinRule Account
// @Description CreateCasbinRule account
// @Accept  json
// @Produce  json
// @Param casbin body model.CasbinRule true "casbin model"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/ [post]
func (h *Handler) CreateCasbinRule(ctx *gin.Context) {
	var casbin model.CasbinRule
	if err := ctx.ShouldBind(&casbin); err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	err := h.usecase.Casbins.CreateCasbinRule(casbin)
	if errors.Is(err, common.NotFound) {
		ctx.JSON(404, gin.H{
			"code":    404,
			"message": "NOT FOUND",
		})
		return
	}
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "CREATE SUCCESS",
	})
}

// DeleteCasbinRule godoc
// @Summary DeleteCasbinRule Account
// @Description DeleteCasbinRule account
// @Accept  json
// @Produce  json
// @Param id  path int  true "id"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id [delete]
func (h *Handler) DeleteCasbinRule(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}

	if err := h.usecase.Casbins.DeleteCasbinRule(id); err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "UPDATE SUCCESS",
	})
}

// UpdateCasbinRulePtype godoc
// @Summary UpdateCasbinRulePtype Account
// @Description UpdateCasbinRulePtype account
// @Accept  json
// @Produce  json
// @Param id  path int  true "id"
// @Param ptype  path string  true "ptype"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id/ptype/:ptype [put]
func (h *Handler) UpdateCasbinRulePtype(ctx *gin.Context) {
	ptype := ctx.Param("ptype")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRulePtype(id, ptype); err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "UPDATE SUCCESS",
	})
}

// UpdateCasbinRuleName godoc
// @Summary UpdateCasbinRuleName Account
// @Description UpdateCasbinRuleName account
// @Accept  json
// @Produce  json
// @Param id  path int  true "id"
// @Param name  path string  true "name"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id/name/:name [put]
func (h *Handler) UpdateCasbinRuleName(ctx *gin.Context) {
	name := ctx.Param("name")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleName(id, name); err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "UPDATE SUCCESS",
	})
}

// UpdateCasbinRuleEndpoint godoc
// @Summary UpdateCasbinRuleEndpoint Account
// @Description UpdateCasbinRuleEndpoint account
// @Accept  json
// @Produce  json
// @Param id  path int  true "id"
// @Param endpoint  path string  true "endpoint"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id/endpoint/:endpoint [put]
func (h *Handler) UpdateCasbinRuleEndpoint(ctx *gin.Context) {
	endpoint := ctx.Param("endpoint")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleEndpoint(id, endpoint); err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "UPDATE SUCCESS",
	})
}

// UpdateCasbinMethod godoc
// @Summary UpdateCasbinMethod Account
// @Description UpdateCasbinMethod account
// @Accept  json
// @Produce  json
// @Param id  path int  true "id"
// @Param method  path string  true "method"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/casbin/:id/method/:method [put]
func (h *Handler) UpdateCasbinMethod(ctx *gin.Context) {
	method := ctx.Param("method")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "BAD REQUEST",
		})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinMethod(id, method); err != nil {
		ctx.JSON(500, gin.H{
			"code":    500,
			"message": "INTERVAL SERVER",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "UPDATE SUCCESS",
	})
}
