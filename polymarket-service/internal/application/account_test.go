package application

import (
	"errors"
	"testing"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func TestDepositTransfersToTreasuryAndCredits(t *testing.T) {
	e := newEnv()
	b, err := e.accounts.Deposit(ctx, "alice", 5_000)
	if err != nil {
		t.Fatal(err)
	}
	tr := e.wallets.transfers[0]
	if b.Available != 5_000 || tr.SenderWalletID != "wallet-user" || tr.ReceiverWalletID != treasury || tr.Amount != 5_000 {
		t.Fatalf("balance %+v transfer %+v", b, tr)
	}
	if len(e.mem.ledger) != 1 || e.mem.ledger[0].Type != model.LedgerDeposit {
		t.Fatalf("ledger %+v", e.mem.ledger)
	}
}

func TestDepositReversesTransferWhenCreditFails(t *testing.T) {
	e := newEnv()
	e.mem.failTx = errors.New("db down")
	if _, err := e.accounts.Deposit(ctx, "alice", 5_000); err == nil {
		t.Fatal("expected failure")
	}
	if len(e.wallets.reversed) != 1 || e.bal("alice").Available != 0 {
		t.Fatalf("reversed=%v balance=%+v", e.wallets.reversed, e.bal("alice"))
	}
}

func TestDepositRejectsWrongCurrency(t *testing.T) {
	e := newEnv()
	e.wallets.wallet.Currency = "VND"
	if _, err := e.accounts.Deposit(ctx, "alice", 100); !errors.Is(err, port.ErrWalletRejected) {
		t.Fatalf("got %v", err)
	}
	if len(e.wallets.transfers) != 0 {
		t.Fatal("no money may move")
	}
}

func TestWithdrawDebitsThenPays(t *testing.T) {
	e := newEnv()
	e.fund("alice", 1_000)
	b, err := e.accounts.Withdraw(ctx, "alice", 400)
	if err != nil || b.Available != 600 {
		t.Fatalf("%+v %v", b, err)
	}
	last := e.wallets.transfers[len(e.wallets.transfers)-1]
	if last.SenderWalletID != treasury || last.ReceiverWalletID != "wallet-user" || last.Amount != 400 {
		t.Fatalf("payout %+v", last)
	}
}

func TestWithdrawRestoresBalanceWhenPayoutFails(t *testing.T) {
	e := newEnv()
	e.fund("alice", 1_000)
	e.wallets.transferErr = port.ErrWalletUnavailable
	if _, err := e.accounts.Withdraw(ctx, "alice", 400); !errors.Is(err, port.ErrWalletUnavailable) {
		t.Fatalf("got %v", err)
	}
	if got := e.bal("alice").Available; got != 1_000 {
		t.Fatalf("balance must be restored, got %d", got)
	}
}

func TestWithdrawCannotTouchLockedFunds(t *testing.T) {
	e := newEnv()
	e.fund("alice", 1_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 50, 15)) // locks 750
	if _, err := e.accounts.Withdraw(ctx, "alice", 500); !errors.Is(err, port.ErrInsufficientBalance) {
		t.Fatalf("got %v", err)
	}
}

func TestPortfolioMarksToMarket(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))
	// Book quotes 64/68 => mid 66, so alice's 10 YES cost 600 and are worth 660.
	e.must(e.limit("bob", model.OutcomeYes, model.Buy, 64, 1))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 32, 1)) // ask yes 68... crosses? bid 64 < 68, rests
	m, _ := e.mem.GetByID(ctx, e.market.ID)
	if m.BestBid != 64 || m.BestAsk != 68 {
		t.Fatalf("top of book %d/%d", m.BestBid, m.BestAsk)
	}

	p, err := e.accounts.Portfolio(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Positions) != 1 {
		t.Fatalf("positions %+v", p.Positions)
	}
	v := p.Positions[0]
	if v.AvgPrice != 60 || v.CurrentPrice != 66 || v.Value != 660 || v.UnrealizedPnL != 60 {
		t.Fatalf("view %+v", v)
	}
	if p.TotalValue != float64(9_400)+660 {
		t.Fatalf("total %v", p.TotalValue)
	}
}

func TestActivityShowsEachPartysOwnView(t *testing.T) {
	e := newEnv()
	e.fund("alice", 10_000)
	e.fund("bob", 10_000)
	e.must(e.limit("alice", model.OutcomeYes, model.Buy, 60, 10))
	e.must(e.limit("bob", model.OutcomeNo, model.Buy, 40, 10))

	items, err := e.accounts.Activity(ctx, "bob", 10)
	if err != nil {
		t.Fatal(err)
	}
	var trade *port.Activity
	for i := range items {
		if items[i].Type == "trade" {
			trade = &items[i]
		}
	}
	if trade == nil || trade.Outcome != model.OutcomeNo || trade.Side != model.Buy || trade.Price != 40 || trade.Amount != 400 {
		t.Fatalf("bob's trade should read as buying NO at 40, got %+v", trade)
	}
}
