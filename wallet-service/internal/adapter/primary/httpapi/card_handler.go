package httpapi

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) issueCard(c *gin.Context) {
	walletID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		IssuerBankID string `json:"issuer_bank_id"`
		Network      string `json:"network" binding:"required"`
		HolderName   string `json:"holder_name" binding:"required"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.IssueCardCommand{
		WalletID:     walletID,
		IssuerBankID: p.required("issuer_bank_id", req.IssuerBankID),
		Network:      model.CardNetwork(req.Network),
		HolderName:   req.HolderName,
	}
	if !p.ok(c) {
		return
	}
	card, err := a.Cards.Issue(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (a *API) listCards(c *gin.Context) {
	walletID, ok := pathID(c, "id")
	if !ok {
		return
	}
	cards, err := a.Cards.ListByWallet(c.Request.Context(), walletID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": cards})
}

func (a *API) getCard(c *gin.Context) {
	a.cardAction(c, a.Cards.Get)
}

func (a *API) blockCard(c *gin.Context) {
	a.cardAction(c, a.Cards.Block)
}

func (a *API) unblockCard(c *gin.Context) {
	a.cardAction(c, a.Cards.Unblock)
}

func (a *API) cardAction(c *gin.Context, fn func(context.Context, uuid.UUID) (*model.Card, error)) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	card, err := fn(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}
