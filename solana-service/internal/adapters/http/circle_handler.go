package http

import (
	"net/http"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/services"
	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
)

type CircleHandler struct {
	walletService      *services.CircleWalletService
	transactionService *services.CircleTransactionService
	tokenService       *services.CircleTokenService
}

func NewCircleHandler(
	walletService *services.CircleWalletService,
	transactionService *services.CircleTransactionService,
	tokenService *services.CircleTokenService,
) *CircleHandler {
	return &CircleHandler{
		walletService:      walletService,
		transactionService: transactionService,
		tokenService:       tokenService,
	}
}

// Wallet Set handlers
type CreateWalletSetRequest struct {
	Name        string             `json:"name" binding:"required"`
	CustodyType domain.CustodyType `json:"custodyType" binding:"required"`
}

func (h *CircleHandler) CreateWalletSet(c *gin.Context) {
	var req CreateWalletSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	walletSet, err := h.walletService.CreateWalletSet(c.Request.Context(), req.Name, req.CustodyType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, walletSet)
}

func (h *CircleHandler) GetWalletSet(c *gin.Context) {
	walletSetID := c.Param("id")

	walletSet, err := h.walletService.GetWalletSet(c.Request.Context(), walletSetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, walletSet)
}

func (h *CircleHandler) ListWalletSets(c *gin.Context) {
	walletSets, err := h.walletService.ListWalletSets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"walletSets": walletSets})
}

// Wallet handlers
type CreateWalletRequest struct {
	WalletSetID string             `json:"walletSetId" binding:"required"`
	Blockchain  string             `json:"blockchain" binding:"required"`
	AccountType domain.AccountType `json:"accountType"`
}

func (h *CircleHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var wallet *domain.CircleWallet
	var err error

	if req.Blockchain == "SOL" {
		wallet, err = h.walletService.CreateSolanaWallet(c.Request.Context(), req.WalletSetID)
	} else {
		wallets, err := h.walletService.CreateMultiChainWallet(c.Request.Context(), req.WalletSetID, []string{req.Blockchain})
		if err == nil && len(wallets) > 0 {
			wallet = wallets[0]
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func (h *CircleHandler) GetWallet(c *gin.Context) {
	walletID := c.Param("id")

	wallet, err := h.walletService.GetWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func (h *CircleHandler) ListWallets(c *gin.Context) {
	walletSetID := c.Query("walletSetId")
	blockchain := c.Query("blockchain")

	wallets, err := h.walletService.ListWallets(c.Request.Context(), walletSetID, blockchain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}

func (h *CircleHandler) GetWalletBalance(c *gin.Context) {
	walletID := c.Param("id")

	balances, err := h.walletService.GetWalletBalance(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

func (h *CircleHandler) GetWalletNFTs(c *gin.Context) {
	walletID := c.Param("id")

	nfts, err := h.walletService.GetWalletNFTs(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nfts": nfts})
}

func (h *CircleHandler) FreezeWallet(c *gin.Context) {
	walletID := c.Param("id")

	wallet, err := h.walletService.FreezeWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func (h *CircleHandler) UnfreezeWallet(c *gin.Context) {
	walletID := c.Param("id")

	wallet, err := h.walletService.UnfreezeWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// Transaction handlers
func (h *CircleHandler) CreateTransfer(c *gin.Context) {
	var req domain.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.CreateTransfer(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *CircleHandler) CreateNFTTransfer(c *gin.Context) {
	var req domain.NFTTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.CreateNFTTransfer(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *CircleHandler) ExecuteContract(c *gin.Context) {
	var req domain.ContractExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.ExecuteContract(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *CircleHandler) GetTransaction(c *gin.Context) {
	txID := c.Param("id")

	tx, err := h.transactionService.GetTransaction(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *CircleHandler) EstimateFee(c *gin.Context) {
	walletID := c.Query("walletId")
	destinationAddress := c.Query("destinationAddress")
	tokenID := c.Query("tokenId")
	amount := c.Query("amount")

	feeEstimate, err := h.transactionService.EstimateFee(c.Request.Context(), walletID, destinationAddress, tokenID, amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feeEstimate)
}

// Bridge handlers
type BridgeSolanaToCircleRequest struct {
	FromSolanaAddress string `json:"fromSolanaAddress" binding:"required"`
	ToCircleWalletID  string `json:"toCircleWalletId" binding:"required"`
	Amount            uint64 `json:"amount" binding:"required"`
	PrivateKey        string `json:"privateKey" binding:"required"`
}

func (h *CircleHandler) BridgeSolanaToCircle(c *gin.Context) {
	var req BridgeSolanaToCircleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	privateKey, err := solana.PrivateKeyFromBase58(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid private key"})
		return
	}

	bridgeTransfer, err := h.transactionService.BridgeSolanaToCircle(
		c.Request.Context(),
		req.FromSolanaAddress,
		req.ToCircleWalletID,
		req.Amount,
		privateKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bridgeTransfer)
}

type BridgeCircleToSolanaRequest struct {
	FromCircleWalletID string `json:"fromCircleWalletId" binding:"required"`
	ToSolanaAddress    string `json:"toSolanaAddress" binding:"required"`
	Amount             string `json:"amount" binding:"required"`
}

func (h *CircleHandler) BridgeCircleToSolana(c *gin.Context) {
	var req BridgeCircleToSolanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bridgeTransfer, err := h.transactionService.BridgeCircleToSolana(
		c.Request.Context(),
		req.FromCircleWalletID,
		req.ToSolanaAddress,
		req.Amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bridgeTransfer)
}

func (h *CircleHandler) GetBridgeTransfer(c *gin.Context) {
	transferID := c.Param("id")

	transfer, err := h.transactionService.GetBridgeTransfer(c.Request.Context(), transferID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// Token handlers
func (h *CircleHandler) GetToken(c *gin.Context) {
	tokenID := c.Param("id")

	token, err := h.tokenService.GetToken(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *CircleHandler) ListTokens(c *gin.Context) {
	blockchain := c.Query("blockchain")

	tokens, err := h.tokenService.ListTokens(c.Request.Context(), blockchain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *CircleHandler) ImportToken(c *gin.Context) {
	var req domain.TokenMetadata
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.tokenService.ImportToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}
