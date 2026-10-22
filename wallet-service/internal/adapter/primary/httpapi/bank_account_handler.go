package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) linkBankAccount(c *gin.Context) {
	userID, ok := pathUserID(c)
	if !ok {
		return
	}
	var req struct {
		BankID        string `json:"bank_id"`
		AccountNumber string `json:"account_number" binding:"required"`
		AccountName   string `json:"account_name" binding:"required"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.LinkBankAccountCommand{
		UserID:        userID,
		BankID:        p.required("bank_id", req.BankID),
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
	}
	if !p.ok(c) {
		return
	}
	account, err := a.BankAccounts.Link(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, account)
}

func (a *API) listBankAccounts(c *gin.Context) {
	userID, ok := pathUserID(c)
	if !ok {
		return
	}
	accounts, err := a.BankAccounts.List(c.Request.Context(), userID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": accounts})
}

func (a *API) removeBankAccount(c *gin.Context) {
	userID, ok := pathUserID(c)
	if !ok {
		return
	}
	accountID, ok := pathID(c, "account_id")
	if !ok {
		return
	}
	if err := a.BankAccounts.Remove(c.Request.Context(), userID, accountID); err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (a *API) setDefaultBankAccount(c *gin.Context) {
	userID, ok := pathUserID(c)
	if !ok {
		return
	}
	accountID, ok := pathID(c, "account_id")
	if !ok {
		return
	}
	account, err := a.BankAccounts.SetDefault(c.Request.Context(), userID, accountID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, account)
}
