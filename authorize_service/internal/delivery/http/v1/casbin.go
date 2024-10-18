package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initCasbinRoutes(api *gin.RouterGroup) {
	group := api.Group("/casbin")

	group.POST("/authorize", h.middleware.AuthorizeControl(), h.Authorize)
	group.GET("/", h.middleware.AuthorizeControl(), h.CasbinRuleAll)
	group.GET("/:id", h.middleware.AuthorizeControl(), h.CasbinRuleById)
	group.POST("/", h.middleware.AuthorizeControl(), h.CreateCasbinRule)
	group.DELETE("/:id", h.middleware.AuthorizeControl(), h.DeleteCasbinRule)
	group.PUT("/:id/ptype/:ptype", h.middleware.AuthorizeControl(), h.UpdateCasbinRulePtype)
	group.PUT("/:id/name/:name", h.middleware.AuthorizeControl(), h.UpdateCasbinRuleName)
	group.PUT("/:id/endpoint/:endpoint", h.middleware.AuthorizeControl(), h.UpdateCasbinRuleEndpoint)
	group.PUT("/:id/method/:method", h.middleware.AuthorizeControl(), h.UpdateCasbinMethod)
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
	log.Info(auth)
	err := h.usecase.Casbins.EnforceCasbin(ctx, auth)
	if errors.Is(err, common.FailedDB) {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.FailedDBServer})
		return
	}
	if errors.Is(err, common.Failedenforcer) {
		Response(ctx, http.StatusUnauthorized, Message{Message: common.Unauthorized})
		return
	}
	if errors.Is(err, common.NotAllow) {
		Response(ctx, http.StatusUnauthorized, Message{Message: common.NotAllowServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.Authorized})
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
	var p pagination.Pagination
	if err := ctx.ShouldBindQuery(&p); err != nil {
		Response(ctx, http.StatusNotFound, Message{Message: common.NotAllowServer})
		return
	}
	casbins, err := h.usecase.Casbins.CasbinRuleAll(ctx, p)
	if errors.Is(err, common.NotFound) {
		Response(ctx, http.StatusNotFound, Message{Message: common.NotAllowServer})
		return
	}
	if err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, casbins)
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
		Response(ctx, http.StatusBadRequest, Message{Message: common.BadRequest})
		return
	}
	casbin, err := h.usecase.Casbins.CasbinRuleById(ctx, id)
	if errors.Is(err, common.NotFound) {
		Response(ctx, http.StatusNotFound, Message{Message: common.NotFoundServer})
		return
	}
	if err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, casbin)
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
		Response(ctx, http.StatusBadRequest, Message{Message: common.BadRequest})
		return
	}
	err := h.usecase.Casbins.CreateCasbinRule(ctx, casbin)
	if errors.Is(err, common.NotFound) {
		Response(ctx, http.StatusNotFound, Message{Message: common.NotFoundServer})
		return
	}
	if err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.CreateSuccess})
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
		Response(ctx, http.StatusBadRequest, Message{Message: common.BadRequest})
		return
	}

	if err := h.usecase.Casbins.DeleteCasbinRule(ctx, id); err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.UpdateSuccess})
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
		Response(ctx, http.StatusNotFound, Message{Message: common.BadRequest})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRulePtype(ctx, id, ptype); err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.UpdateSuccess})
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
		Response(ctx, http.StatusNotFound, Message{Message: common.BadRequest})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleName(ctx, id, name); err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.UpdateSuccess})
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
		Response(ctx, http.StatusNotFound, Message{Message: common.BadRequest})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleEndpoint(ctx, id, endpoint); err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.UpdateSuccess})
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
		Response(ctx, http.StatusNotFound, Message{Message: common.BadRequest})
		return
	}
	if err := h.usecase.Casbins.UpdateCasbinMethod(ctx, id, method); err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: common.InternalServer})
		return
	}
	Response(ctx, http.StatusOK, Message{Message: common.UpdateSuccess})
}
