package http

import (
	stdhttp "net/http"

	"github.com/JIeeiroSst/voucher-service/internal/adapters/inbound/http/middleware"
	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/gin-gonic/gin"
)

func forbid(c *gin.Context) {
	c.JSON(stdhttp.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "you do not have access to this resource"}})
}

func callerIsUser(c *gin.Context, id string) bool {
	caller := middleware.CallerID(c)
	return caller != "" && caller == id
}

func canAccessOrder(c *gin.Context, o *order.Order) bool {
	return middleware.IsAdmin(c) || (o.BuyerType == order.BuyerTypeRetail && callerIsUser(c, o.BuyerID))
}

func canAccessVoucher(c *gin.Context, v *voucher.Voucher) bool {
	return middleware.IsAdmin(c) || (v.OwnerType == voucher.OwnerTypeUser && v.OwnerID != nil && callerIsUser(c, *v.OwnerID))
}
