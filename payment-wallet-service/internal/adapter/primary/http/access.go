package http

import (
	"context"

	"github.com/JIeeiroSst/payment-wallet-service/pkg/authn"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ownsWallet(ctx context.Context, p authn.Principal, walletID string) (bool, error) {
	if p.IsAdmin() {
		return true, nil
	}
	w, err := h.wallets.GetWallet(ctx, walletID)
	if err != nil {
		return false, err
	}
	return p.Owns(w.UserID), nil
}

func (h *Handler) guard(c *gin.Context, check func(context.Context, authn.Principal) (bool, error)) {
	p := principal(c)
	ok, err := check(c.Request.Context(), p)
	if err != nil {
		writeError(c, err)
		c.Abort()
		return
	}
	if !ok {
		abortAuth(c, authn.ErrForbidden)
		return
	}
	c.Next()
}

func (h *Handler) walletParam(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
			return h.ownsWallet(ctx, p, c.Param(param))
		})
	}
}

func (h *Handler) walletBody(field string, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
			id := bodyString(c, field)
			if id == "" {
				return !required || p.IsAdmin(), nil
			}
			return h.ownsWallet(ctx, p, id)
		})
	}
}

func (h *Handler) walletQuery(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
			return h.ownsWallet(ctx, p, c.Query(field))
		})
	}
}

func userParam(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !principal(c).Owns(c.Param(param)) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func userQuery(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !principal(c).Owns(c.Query(field)) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func forceUserBody(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if p := principal(c); !p.IsAdmin() {
			setBodyField(c, field, p.UserID)
		}
		c.Next()
	}
}

func (h *Handler) pocketParam(c *gin.Context) {
	h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
		if p.IsAdmin() {
			return true, nil
		}
		pocket, err := h.pockets.GetPocket(ctx, c.Param("pocketId"))
		if err != nil {
			return false, err
		}
		return h.ownsWallet(ctx, p, pocket.WalletID)
	})
}

func (h *Handler) transactionParam(c *gin.Context) {
	h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
		if p.IsAdmin() {
			return true, nil
		}
		txn, err := h.transactions.GetTransaction(ctx, c.Param("id"))
		if err != nil {
			return false, err
		}
		return h.ownsWallet(ctx, p, txn.WalletID)
	})
}

func (h *Handler) transferParam(c *gin.Context) {
	h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
		if p.IsAdmin() {
			return true, nil
		}
		t, err := h.transactions.GetTransfer(ctx, c.Param("id"))
		if err != nil {
			return false, err
		}
		if ok, err := h.ownsWallet(ctx, p, t.SenderWalletID); ok || err != nil {
			return ok, err
		}
		return h.ownsWallet(ctx, p, t.ReceiverWalletID)
	})
}

func (h *Handler) requesterParam(c *gin.Context) {
	h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
		if p.IsAdmin() {
			return true, nil
		}
		r, err := h.paymentRequests.GetPaymentRequest(ctx, c.Param("id"))
		if err != nil {
			return false, err
		}
		return h.ownsWallet(ctx, p, r.RequesterWalletID)
	})
}

func (h *Handler) paymentMethodParam(c *gin.Context) {
	h.guard(c, func(ctx context.Context, p authn.Principal) (bool, error) {
		if p.IsAdmin() {
			return true, nil
		}
		m, err := h.paymentMethods.GetPaymentMethod(ctx, c.Param("id"))
		if err != nil {
			return false, err
		}
		return p.Owns(m.UserID), nil
	})
}
