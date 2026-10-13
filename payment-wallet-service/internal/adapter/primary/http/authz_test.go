package http

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

type fakeWallets struct {
	port.WalletUsecase
	createdFor string
}

func (f *fakeWallets) GetWallet(_ context.Context, id string) (*model.Wallet, error) {
	switch id {
	case "w7":
		return &model.Wallet{WalletID: "w7", UserID: "7"}, nil
	case "w8":
		return &model.Wallet{WalletID: "w8", UserID: "8"}, nil
	}
	return nil, port.ErrNotFound
}
func (f *fakeWallets) CreateWallet(_ context.Context, userID, currency string) (*model.Wallet, error) {
	f.createdFor = userID
	return &model.Wallet{WalletID: "new", UserID: userID}, nil
}
func (f *fakeWallets) GetWalletByUser(_ context.Context, userID string) (*model.Wallet, error) {
	return &model.Wallet{UserID: userID}, nil
}

type fakeTransactions struct {
	port.TransactionUsecase
	transferred bool
}

func (f *fakeTransactions) Transfer(_ context.Context, from, to string, amount int64, ref, desc string) (*model.Transfer, error) {
	f.transferred = true
	return &model.Transfer{SenderWalletID: from, ReceiverWalletID: to}, nil
}
func (f *fakeTransactions) ListTransactions(context.Context, string, int, int) ([]model.Transaction, error) {
	return nil, nil
}

type fakePocketSvc struct{ port.PocketUsecase }

func (fakePocketSvc) GetPocket(_ context.Context, id string) (*model.Pocket, error) {
	return &model.Pocket{WalletID: map[string]string{"p7": "w7", "p8": "w8"}[id]}, nil
}
func (fakePocketSvc) DepositToPocket(context.Context, string, int64) (*model.Pocket, error) {
	return &model.Pocket{}, nil
}

func TestWalletAuthorization(t *testing.T) {
	fakeUserService(t)
	wallets := &fakeWallets{}
	txns := &fakeTransactions{}
	router := NewRouter(NewHandler(wallets, txns, nil, fakePocketSvc{}, nil))
	user := testToken("7", "user")
	admin := testToken("1", "admin")

	cases := []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous", authCall{method: "GET", path: "/api/v1/wallets/w7"}, 401},
		{"own wallet", authCall{method: "GET", path: "/api/v1/wallets/w7", bearer: user}, 200},
		{"foreign wallet", authCall{method: "GET", path: "/api/v1/wallets/w8", bearer: user}, 403},
		{"admin foreign wallet", authCall{method: "GET", path: "/api/v1/wallets/w8", bearer: admin}, 200},
		{"foreign by user", authCall{method: "GET", path: "/api/v1/wallets/user/8", bearer: user}, 403},
		{"user deposit blocked", authCall{method: "POST", path: "/api/v1/wallets/w7/deposit", body: `{"amount":10}`, bearer: user}, 403},
		{"user freeze blocked", authCall{method: "POST", path: "/api/v1/wallets/w7/freeze", body: `{}`, bearer: user}, 403},
		{"foreign transactions", authCall{method: "GET", path: "/api/v1/wallets/w8/transactions", bearer: user}, 403},
		{"own transactions", authCall{method: "GET", path: "/api/v1/wallets/w7/transactions", bearer: user}, 200},
		{"foreign pocket", authCall{method: "POST", path: "/api/v1/pockets/p8/deposit", body: `{"amount":1}`, bearer: user}, 403},
		{"own pocket", authCall{method: "POST", path: "/api/v1/pockets/p7/deposit", body: `{"amount":1}`, bearer: user}, 200},
		{"transfer from foreign wallet", authCall{method: "POST", path: "/api/v1/transfers", body: `{"sender_wallet_id":"w8","receiver_wallet_id":"w7","amount":5,"reference_id":"r"}`, bearer: user}, 403},
		{"internal transfer", authCall{method: "POST", path: "/api/v1/transfers", body: `{"sender_wallet_id":"w8","receiver_wallet_id":"w7","amount":5,"reference_id":"r"}`, internal: testInternalToken}, 201},
	}
	for _, c := range cases {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
	txns.transferred = false
	if code, body := doAuth(router, authCall{method: "POST", path: "/api/v1/transfers", body: `{"sender_wallet_id":"w7","receiver_wallet_id":"w8","amount":5,"reference_id":"r"}`, bearer: user}); code != 201 || !txns.transferred {
		t.Fatalf("own transfer: %d %s", code, body)
	}
	if code, _ := doAuth(router, authCall{method: "POST", path: "/api/v1/wallets", body: `{"user_id":"8","currency":"VND"}`, bearer: user}); code != 201 || wallets.createdFor != "7" {
		t.Fatalf("create wallet forced owner: code=%d owner=%s", code, wallets.createdFor)
	}
}
