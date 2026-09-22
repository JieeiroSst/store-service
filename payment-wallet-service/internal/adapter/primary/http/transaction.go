package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type moneyMovementRequest struct {
	Amount      int64  `json:"amount" binding:"required"`
	ReferenceID string `json:"reference_id"`
	Description string `json:"description"`
}

func (h *Handler) Deposit(c *gin.Context) {
	var req moneyMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactions.Deposit(c.Request.Context(), c.Param("id"), req.Amount, req.ReferenceID, req.Description)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, txn)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var req moneyMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactions.Withdraw(c.Request.Context(), c.Param("id"), req.Amount, req.ReferenceID, req.Description)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, txn)
}

type transferRequest struct {
	SenderWalletID   string `json:"sender_wallet_id" binding:"required"`
	ReceiverWalletID string `json:"receiver_wallet_id" binding:"required"`
	Amount           int64  `json:"amount" binding:"required"`
	ReferenceID      string `json:"reference_id"`
	Description      string `json:"description"`
}

func (h *Handler) Transfer(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.transactions.Transfer(c.Request.Context(), req.SenderWalletID, req.ReceiverWalletID, req.Amount, req.ReferenceID, req.Description)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	txn, err := h.transactions.GetTransaction(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, txn)
}

func (h *Handler) ListTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	txns, err := h.transactions.ListTransactions(c.Request.Context(), c.Param("id"), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, txns)
}

func (h *Handler) GetTransfer(c *gin.Context) {
	transfer, err := h.transactions.GetTransfer(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, transfer)
}

type reverseRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *Handler) ReverseTransaction(c *gin.Context) {
	var req reverseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactions.Reverse(c.Request.Context(), c.Param("id"), req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, txn)
}

func (h *Handler) ReverseTransfer(c *gin.Context) {
	var req reverseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.transactions.ReverseTransfer(c.Request.Context(), c.Param("id"), req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, transfer)
}

// Statement exports a wallet's ledger for [from, to) as CSV -
// from/to are YYYY-MM-DD, defaulting to the last 30 days ending today.
func (h *Handler) Statement(c *gin.Context) {
	to := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
	from := to.Add(-30 * 24 * time.Hour)
	if v := c.Query("to"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, expected YYYY-MM-DD"})
			return
		}
		to = parsed.Add(24 * time.Hour)
	}
	if v := c.Query("from"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, expected YYYY-MM-DD"})
			return
		}
		from = parsed
	}

	walletID := c.Param("id")
	txns, err := h.transactions.Statement(c.Request.Context(), walletID, from, to)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-statement.csv", walletID))

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"transaction_id", "type", "amount", "currency", "status", "reference_id", "description", "created_at"})
	for _, t := range txns {
		_ = w.Write([]string{
			t.TransactionID,
			string(t.Type),
			strconv.FormatInt(t.Amount, 10),
			t.Currency,
			string(t.Status),
			t.ReferenceID,
			t.Description,
			t.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
}
