package application

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func TestPayMerchantAndRefundInParts(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "300", true)

	pay, err := h.payments.Pay(h.ctx, port.PayCommand{
		IdempotencyKey: "order-1", UserID: alice, WalletID: w.ID, MerchantID: h.merchant.ID,
		Amount: dec("100"), Pin: testPin, Metadata: map[string]string{"order_id": "A1"},
	})
	mustNoErr(t, err)
	if pay.MerchantID.UUID != h.merchant.ID || pay.AcquirerBankID.UUID != h.acquirer.ID || pay.Metadata["order_id"] != "A1" {
		t.Fatalf("payment not linked to merchant: %+v", pay)
	}
	mustEqualDec(t, "after pay", h.reload(w.ID).Balance, "200")

	r1, err := h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf-1", TransactionID: pay.ID, Amount: dec("30"), Reason: "missing item"})
	mustNoErr(t, err)
	if r1.ReferenceID.UUID != pay.ID {
		t.Fatalf("refund must reference payment")
	}
	orig, _ := txRepo{db: h.db}.GetByID(h.ctx, pay.ID)
	if orig.Status != model.TransactionStatusPartiallyRefunded {
		t.Fatalf("expected PARTIALLY_REFUNDED, got %s", orig.Status)
	}

	_, err = h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf-2", TransactionID: pay.ID, Amount: dec("80")})
	mustErr(t, err, model.ErrRefundExceeds)

	r2, err := h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf-3", TransactionID: pay.ID})
	mustNoErr(t, err)
	mustEqualDec(t, "remaining refund", r2.Amount, "70")
	orig, _ = txRepo{db: h.db}.GetByID(h.ctx, pay.ID)
	if orig.Status != model.TransactionStatusRefunded {
		t.Fatalf("expected REFUNDED, got %s", orig.Status)
	}
	mustEqualDec(t, "after refunds", h.reload(w.ID).Balance, "300")

	_, err = h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf-4", TransactionID: pay.ID})
	mustErr(t, err, model.ErrNotRefundable)

	replay, err := h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf-1", TransactionID: pay.ID, Amount: dec("30")})
	mustNoErr(t, err)
	if replay.ID != r1.ID {
		t.Fatalf("refund replay must return original refund")
	}
}

func TestRefundRejectsNonPayments(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyUSD, "100", true)
	h.newWallet(bob, model.CurrencyUSD, "0", false)
	tr, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "t", UserID: alice, FromWalletID: src.ID, ToUserID: bob, Amount: dec("10"), Pin: testPin})
	mustNoErr(t, err)

	_, err = h.payments.Refund(h.ctx, port.RefundCommand{IdempotencyKey: "rf", TransactionID: tr.ID})
	mustErr(t, err, model.ErrNotRefundable)
}

func TestCardAuthorizeCaptureAndVoid(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "500", false)
	card, err := h.cards.Issue(h.ctx, port.IssueCardCommand{WalletID: w.ID, IssuerBankID: h.issuer.ID, Network: "visa", HolderName: "Alice Nguyen"})
	mustNoErr(t, err)
	if card.CardNumber[:4] != "****" || card.Network != model.CardNetworkVisa || card.HolderName != "ALICE NGUYEN" {
		t.Fatalf("unexpected issued card: %+v", card)
	}

	auth, err := h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "pos-1", CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec("200"), Currency: "usd"})
	mustNoErr(t, err)
	mustEqualDec(t, "interchange fee", auth.Fee, "3")
	held := h.reload(w.ID)
	mustEqualDec(t, "balance unchanged on auth", held.Balance, "500")
	mustEqualDec(t, "held", held.FrozenAmount, "200")

	_, err = h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "pos-2", CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec("301"), Currency: "USD"})
	mustErr(t, err, model.ErrInsufficientFunds)

	captured, err := h.cardPay.Capture(h.ctx, auth.ID)
	mustNoErr(t, err)
	if captured.Status != model.TransactionStatusCaptured {
		t.Fatalf("expected CAPTURED, got %s", captured.Status)
	}
	after := h.reload(w.ID)
	mustEqualDec(t, "balance after capture", after.Balance, "300")
	mustEqualDec(t, "held after capture", after.FrozenAmount, "0")

	_, err = h.cardPay.Capture(h.ctx, auth.ID)
	mustErr(t, err, model.ErrTransactionState)

	auth2, err := h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "pos-3", CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec("50"), Currency: "USD"})
	mustNoErr(t, err)
	_, err = h.cardPay.Void(h.ctx, auth2.ID)
	mustNoErr(t, err)
	mustEqualDec(t, "held after void", h.reload(w.ID).FrozenAmount, "0")

	_, err = h.cards.Block(h.ctx, card.ID)
	mustNoErr(t, err)
	_, err = h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "pos-4", CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec("1"), Currency: "USD"})
	mustErr(t, err, model.ErrCardBlocked)
	_, err = h.cards.Unblock(h.ctx, card.ID)
	mustNoErr(t, err)

	_, err = h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "pos-5", CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec("1"), Currency: "EUR"})
	mustErr(t, err, model.ErrCurrencyMismatch)
}

func TestSettlementBatchClearAndSettle(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "1000", false)
	card, err := h.cards.Issue(h.ctx, port.IssueCardCommand{WalletID: w.ID, IssuerBankID: h.issuer.ID, Network: model.CardNetworkVisa, HolderName: "Alice"})
	mustNoErr(t, err)
	for i, amount := range []string{"100", "200"} {
		auth, err := h.cardPay.Authorize(h.ctx, port.AuthorizeCommand{IdempotencyKey: "s-" + amount, CardID: card.ID, MerchantID: h.merchant.ID, Amount: dec(amount), Currency: "USD"})
		mustNoErr(t, err)
		_, err = h.cardPay.Capture(h.ctx, auth.ID)
		mustNoErr(t, err)
		h.advance(time.Duration(i+1) * time.Minute)
	}

	batch, err := h.settle.CreateBatch(h.ctx, h.merchant.ID)
	mustNoErr(t, err)
	if batch.TxnCount != 2 || batch.Status != model.BatchStatusOpen {
		t.Fatalf("unexpected batch: %+v", batch)
	}
	mustEqualDec(t, "batch total", batch.TotalAmount, "300")
	mustEqualDec(t, "batch fee", batch.TotalFee, "4.5")

	_, err = h.settle.CreateBatch(h.ctx, h.merchant.ID)
	mustErr(t, err, model.ErrNothingToSettle)

	mustErr(t, h.settle.Settle(h.ctx, batch.ID), model.ErrBatchState)

	records, err := h.settle.Clear(h.ctx, batch.ID)
	mustNoErr(t, err)
	if len(records) != 1 {
		t.Fatalf("expected one netted record, got %d", len(records))
	}
	mustEqualDec(t, "net amount", records[0].NetAmount, "295.5")
	if records[0].CardNetwork != model.CardNetworkVisa {
		t.Fatalf("clearing must use the card network of the transactions")
	}
	again, err := h.settle.Clear(h.ctx, batch.ID)
	mustNoErr(t, err)
	if len(again) != 1 || again[0].ID != records[0].ID {
		t.Fatalf("clearing must be idempotent")
	}

	mustNoErr(t, h.settle.Settle(h.ctx, batch.ID))
	mustNoErr(t, h.settle.Settle(h.ctx, batch.ID))
	for _, tx := range h.db.txs {
		if tx.BatchID.Valid && tx.Status != model.TransactionStatusSettled {
			t.Fatalf("batched tx not settled: %s", tx.Status)
		}
	}
}
