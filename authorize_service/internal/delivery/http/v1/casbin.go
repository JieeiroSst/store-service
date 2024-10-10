package v1

import (
	"errors"
	"strconv"

	"github.com/JIeeiroSst/utils/response"
	"github.com/JIeeiroSst/utils/trace_id"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
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
	newCtx := trace_id.TracerID("")
	username := ctx.Query("username")
	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	auth := model.CasbinAuth{
		Sub: username,
		Obj: path,
		Act: method,
	}
	log.Info(auth)
	err := h.usecase.Casbins.EnforceCasbin(newCtx, auth)
	if errors.Is(err, common.FailedDB) {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.FailedDBServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if errors.Is(err, common.Failedenforcer) {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.Unauthorized,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if errors.Is(err, common.NotAllow) {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.NotAllowServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.Authorized,
		Error:   false,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	casbins, err := h.usecase.Casbins.CasbinRuleAll(newCtx)
	if errors.Is(err, common.NotFound) {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.NotAllowServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Error:   false,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		Data:    casbins,
	})

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
	newCtx := trace_id.TracerID("")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	casbin, err := h.usecase.Casbins.CasbinRuleById(newCtx, id)
	if errors.Is(err, common.NotFound) {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.NotFoundServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		Data:    casbin,
	})
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
	newCtx := trace_id.TracerID("")
	var casbin model.CasbinRule
	if err := ctx.ShouldBind(&casbin); err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	err := h.usecase.Casbins.CreateCasbinRule(newCtx, casbin)
	if errors.Is(err, common.NotFound) {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.NotFoundServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.CreateSuccess,
		Error:   false,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}

	if err := h.usecase.Casbins.DeleteCasbinRule(newCtx, id); err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.UpdateSuccess,
		Error:   false,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	ptype := ctx.Param("ptype")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err := h.usecase.Casbins.UpdateCasbinRulePtype(newCtx, id, ptype); err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.UpdateSuccess,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	name := ctx.Param("name")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleName(newCtx, id, name); err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.UpdateSuccess,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	endpoint := ctx.Param("endpoint")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err := h.usecase.Casbins.UpdateCasbinRuleEndpoint(newCtx, id, endpoint); err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.UpdateSuccess,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
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
	newCtx := trace_id.TracerID("")
	method := ctx.Param("method")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: common.BadRequest,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err := h.usecase.Casbins.UpdateCasbinMethod(newCtx, id, method); err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: common.InternalServer,
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.UpdateSuccess,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
	})
}
