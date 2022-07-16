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
