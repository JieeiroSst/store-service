package v1

import (
	"strconv"

	"github.com/JIeeiroSst/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initTransaction(api *gin.RouterGroup) {
	_ = api.Group("/transaction")
}

func (h *Handler) GetTransactions(ctx *gin.Context) {
	transactionID, err := strconv.Atoi(ctx.Query("transaction_id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Error:   true,
			Message: err.Error(),
		})
	}

	transactionResponse, err := h.usecase.Transactions.GetTransactions(ctx, transactionID)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Error:   true,
			Message: err.Error(),
		})
	}

	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: err.Error(),
		Data:    transactionResponse,
	})
}
