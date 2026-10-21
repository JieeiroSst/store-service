package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) createWallet(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		Currency string `json:"currency" binding:"required"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	userID := p.required("user_id", req.UserID)
	if !p.ok(c) {
		return
	}
	w, err := a.Wallets.Create(c.Request.Context(), userID, model.Currency(req.Currency))
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (a *API) getWallet(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	w, err := a.Wallets.Get(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (a *API) listUserWallets(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}
	wallets, err := a.Wallets.ListByUser(c.Request.Context(), userID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": wallets})
}

func (a *API) getUserWallet(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}
	w, err := a.Wallets.GetByUserAndCurrency(c.Request.Context(), userID, model.Currency(c.Param("currency")))
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (a *API) getBalance(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	b, err := a.Wallets.Balance(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (a *API) getLimits(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	usage, err := a.Wallets.Limits(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, usage)
}

func (a *API) credit(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Amount         decimal.Decimal `json:"amount"`
		Description    string          `json:"description"`
		IdempotencyKey string          `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	tx, err := a.Wallets.Credit(c.Request.Context(), port.CreditCommand{
		WalletID:       id,
		Amount:         req.Amount,
		Description:    req.Description,
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
	})
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (a *API) freezeWallet(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	w, err := a.Wallets.Freeze(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "frozen", "wallet": w})
}

func (a *API) unfreezeWallet(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	w, err := a.Wallets.Unfreeze(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "active", "wallet": w})
}

func (a *API) closeWallet(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	userID := p.required("user_id", req.UserID)
	if !p.ok(c) {
		return
	}
	w, err := a.Wallets.Close(c.Request.Context(), id, userID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (a *API) changeTier(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Tier string `json:"tier" binding:"required"`
	}
	if !bind(c, &req) {
		return
	}
	w, err := a.Wallets.ChangeTier(c.Request.Context(), id, model.Tier(req.Tier))
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (a *API) setPin(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID     string `json:"user_id"`
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin" binding:"required"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	userID := p.required("user_id", req.UserID)
	if !p.ok(c) {
		return
	}
	err := a.Wallets.SetPin(c.Request.Context(), port.SetPinCommand{
		WalletID:   id,
		UserID:     userID,
		CurrentPin: req.CurrentPin,
		NewPin:     req.NewPin,
	})
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "pin_updated"})
}

func (a *API) listTransactions(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	from, to, ok := timeRange(c)
	if !ok {
		return
	}
	pg, size := pagination(c)
	filter := model.TransactionFilter{WalletID: id, From: from, To: to, Page: pg, PageSize: size}
	for _, t := range csv(c.Query("type")) {
		filter.Types = append(filter.Types, model.TransactionType(t))
	}
	for _, s := range csv(c.Query("status")) {
		filter.Statuses = append(filter.Statuses, model.TransactionStatus(s))
	}
	items, total, err := a.History.Transactions(c.Request.Context(), filter)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page[*model.Transaction]{Items: items, Total: total, Page: pg, PageSize: size})
}

func (a *API) listLedger(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	from, to, ok := timeRange(c)
	if !ok {
		return
	}
	pg, size := pagination(c)
	items, total, err := a.History.Ledger(c.Request.Context(), id, from, to, pg, size)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page[*model.LedgerEntry]{Items: items, Total: total, Page: pg, PageSize: size})
}

func (a *API) statement(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	from, to, ok := timeRange(c)
	if !ok {
		return
	}
	if from == nil || to == nil {
		badRequest(c, "from and to are required")
		return
	}
	st, err := a.History.Statement(c.Request.Context(), id, *from, *to)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}
