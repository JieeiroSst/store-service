package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1")
	{
		coupons := api.Group("/coupons")
		{
			coupons.POST("", h.CreateCoupon)
			coupons.GET("", h.ListCoupons)
			coupons.GET("/code/:code", h.GetCouponByCode)
			coupons.GET("/:id", h.GetCoupon)
			coupons.PUT("/:id", h.UpdateCoupon)
			coupons.DELETE("/:id", h.DeleteCoupon)

			coupons.GET("/:id/restrictions", h.ListRestrictions)
			coupons.POST("/:id/restrictions", h.CreateRestriction)
			coupons.GET("/:id/usages", h.ListUsagesByCoupon)
			coupons.GET("/:id/user-coupons", h.ListUserCouponsByCoupon)

			coupons.POST("/validate", h.ValidateCoupon)
			coupons.POST("/apply", h.ApplyCoupon)
		}

		api.PUT("/restrictions/:id", h.UpdateRestriction)
		api.DELETE("/restrictions/:id", h.DeleteRestriction)

		userCoupons := api.Group("/user-coupons")
		{
			userCoupons.POST("", h.AssignCoupon)
			userCoupons.GET("", h.ListUserCouponsByUser) // ?user_id=
			userCoupons.POST("/:id/use", h.UseCoupon)
			userCoupons.POST("/:id/unuse", h.UnuseCoupon)
			userCoupons.DELETE("/:id", h.DeleteUserCoupon)
		}

		api.GET("/users/:user_id/coupon-usages", h.ListUsagesByUser)
	}

	return engine
}
