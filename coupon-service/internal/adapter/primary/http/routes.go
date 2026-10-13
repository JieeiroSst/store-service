package http

import (
	"net/http"

	"github.com/JIeeiroSst/coupon-service/pkg/authn"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1", authenticate(authn.FromEnv()), requireUser())
	{
		coupons := api.Group("/coupons")
		{
			coupons.POST("", requireAdmin(), h.CreateCoupon)
			coupons.GET("", requireAdmin(), h.ListCoupons)
			coupons.GET("/code/:code", h.GetCouponByCode)
			coupons.GET("/:id", requireAdmin(), h.GetCoupon)
			coupons.PUT("/:id", requireAdmin(), h.UpdateCoupon)
			coupons.DELETE("/:id", requireAdmin(), h.DeleteCoupon)

			coupons.GET("/:id/restrictions", requireAdmin(), h.ListRestrictions)
			coupons.POST("/:id/restrictions", requireAdmin(), h.CreateRestriction)
			coupons.GET("/:id/usages", requireAdmin(), h.ListUsagesByCoupon)
			coupons.GET("/:id/user-coupons", requireAdmin(), h.ListUserCouponsByCoupon)

			coupons.POST("/validate", forceOwnerBody("user_id", true), h.ValidateCoupon)
			coupons.POST("/apply", forceOwnerBody("user_id", true), h.ApplyCoupon)
		}

		api.PUT("/restrictions/:id", requireAdmin(), h.UpdateRestriction)
		api.DELETE("/restrictions/:id", requireAdmin(), h.DeleteRestriction)

		userCoupons := api.Group("/user-coupons")
		{
			userCoupons.POST("", requireAdmin(), h.AssignCoupon)
			userCoupons.GET("", ownerQuery("user_id"), h.ListUserCouponsByUser)
			userCoupons.POST("/:id/use", requireAdmin(), h.UseCoupon)
			userCoupons.POST("/:id/unuse", requireAdmin(), h.UnuseCoupon)
			userCoupons.DELETE("/:id", requireAdmin(), h.DeleteUserCoupon)
		}

		api.GET("/users/:user_id/coupon-usages", ownerParam("user_id"), h.ListUsagesByUser)
	}

	return engine
}
