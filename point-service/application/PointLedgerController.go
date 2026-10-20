package application

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/point-service/domain/service"
	"github.com/gin-gonic/gin"
)

type PointLedgerController struct {
	ledger service.PointLedgerService
}

func InitPointLedgerRouter(router *gin.Engine, ledger service.PointLedgerService) {
	registerPointLedgerRoutes(router, &PointLedgerController{ledger: ledger})
}

func registerPointLedgerRoutes(router gin.IRouter, c *PointLedgerController) {
	group := router.Group("/api/v1/users/:userId")
	group.GET("/points", c.GetBalanceHandler)
	group.GET("/point-transactions", c.ListTransactionsHandler)
}

func (c *PointLedgerController) GetBalanceHandler(ctx *gin.Context) {
	resp, err := c.ledger.Balance(ctx, ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func (c *PointLedgerController) ListTransactionsHandler(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}
	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "offset must be an integer"})
		return
	}
	resp, err := c.ledger.Transactions(ctx, ctx.Param("userId"), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}
