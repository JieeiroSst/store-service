package http

import (
	"net/http"

	"github.com/JIeeiroSst/solana-service/internal/core/services"
	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	accountService     *services.AccountService
	transactionService *services.TransactionService
	programService     *services.ProgramService
}

func NewHandler(
	accountService *services.AccountService,
	transactionService *services.TransactionService,
	programService *services.ProgramService,
) *Handler {
	return &Handler{
		accountService:     accountService,
		transactionService: transactionService,
		programService:     programService,
	}
}

// GetAccount handles GET /accounts/:address
func (h *Handler) GetAccount(c *gin.Context) {
	address := c.Param("address")

	account, err := h.accountService.GetAccountInfo(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetBalance handles GET /accounts/:address/balance
func (h *Handler) GetBalance(c *gin.Context) {
	address := c.Param("address")

	balance, err := h.accountService.GetBalance(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

// GetTransaction handles GET /transactions/:signature
func (h *Handler) GetTransaction(c *gin.Context) {
	signature := c.Param("signature")

	tx, err := h.transactionService.GetTransaction(c.Request.Context(), signature)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// EstimateFee handles POST /transactions/estimate-fee
func (h *Handler) EstimateFee(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	from, _ := solana.PublicKeyFromBase58(req.From)
	to, _ := solana.PublicKeyFromBase58(req.To)

	feeInfo, err := h.transactionService.EstimateFee(c.Request.Context(), from, to, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feeInfo)
}

// GetProgram handles GET /programs/:id
func (h *Handler) GetProgram(c *gin.Context) {
	programID := c.Param("id")

	program, err := h.programService.GetProgram(c.Request.Context(), programID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, program)
}

// FindPDA handles POST /programs/pda
func (h *Handler) FindPDA(c *gin.Context) {
	var req PDARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pda, err := h.programService.CreatePDAExample(
		c.Request.Context(),
		req.ProgramID,
		req.UserPubkey,
		req.Identifier,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": pda.Address.String(),
		"bump":    pda.Bump,
	})
}
