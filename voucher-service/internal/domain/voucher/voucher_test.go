package voucher_test

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/stretchr/testify/require"
)

func newTestVoucher(t *testing.T) *voucher.Voucher {
	t.Helper()
	ref := shared.ProductRef{
		MerchantID:   shared.NewMerchantID(),
		SKU:          "SKU-1",
		Denomination: shared.NewMoney(100000, "VND"),
	}
	v, err := voucher.NewVoucher(ref.MerchantID, ref, time.Now())
	require.NoError(t, err)
	require.Equal(t, voucher.StatusCreated, v.Status)
	require.Equal(t, 1, v.Version)
	return v
}

func TestVoucher_FullLifecycle_Succeeds(t *testing.T) {
	v := newTestVoucher(t)
	now := time.Now()

	require.NoError(t, v.Issue(shared.VoucherCode{Code: "CODE1", PIN: "1234"}, nil, now))
	require.Equal(t, voucher.StatusIssued, v.Status)
	require.Equal(t, 2, v.Version)

	require.NoError(t, v.Activate(voucher.OwnerTypeUser, "user-1", now))
	require.Equal(t, voucher.StatusActive, v.Status)
	require.Equal(t, 3, v.Version)

	require.NoError(t, v.ValidatePIN("1234"))
	require.Error(t, v.ValidatePIN("wrong"))

	amount := shared.NewMoney(100000, "VND")
	require.NoError(t, v.Redeem(amount, "txn-ref-1", now))
	require.Equal(t, voucher.StatusRedeemed, v.Status)
	require.Equal(t, 4, v.Version)
	require.NotNil(t, v.RedeemedAmount)
	require.True(t, v.RedeemedAmount.Equal(amount))
}

func TestVoucher_Redeem_FromCreated_IsRejected(t *testing.T) {
	v := newTestVoucher(t)
	err := v.Redeem(shared.NewMoney(1, "VND"), "", time.Now())
	require.ErrorIs(t, err, voucher.ErrInvalidTransition)
}

func TestVoucher_Redeem_Twice_IsRejected(t *testing.T) {
	v := newTestVoucher(t)
	now := time.Now()
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "CODE1"}, nil, now))
	require.NoError(t, v.Activate(voucher.OwnerTypeUser, "user-1", now))
	amount := shared.NewMoney(100000, "VND")
	require.NoError(t, v.Redeem(amount, "", now))

	err := v.Redeem(amount, "", now)
	require.ErrorIs(t, err, voucher.ErrAlreadyRedeemed)
}

func TestVoucher_Redeem_Expired_IsRejected(t *testing.T) {
	v := newTestVoucher(t)
	now := time.Now()
	past := now.Add(-time.Hour)
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "CODE1"}, &past, now))
	require.NoError(t, v.Activate(voucher.OwnerTypeUser, "user-1", now))

	err := v.Redeem(shared.NewMoney(1, "VND"), "", now)
	require.ErrorIs(t, err, voucher.ErrVoucherExpired)
}

func TestVoucher_Revoke_FromEachRevocableState_Succeeds(t *testing.T) {
	now := time.Now()

	t.Run("from created", func(t *testing.T) {
		v := newTestVoucher(t)
		require.NoError(t, v.Revoke("fraud", now))
		require.Equal(t, voucher.StatusRevoked, v.Status)
	})

	t.Run("from issued", func(t *testing.T) {
		v := newTestVoucher(t)
		require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, nil, now))
		require.NoError(t, v.Revoke("fraud", now))
		require.Equal(t, voucher.StatusRevoked, v.Status)
	})

	t.Run("from active", func(t *testing.T) {
		v := newTestVoucher(t)
		require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, nil, now))
		require.NoError(t, v.Activate(voucher.OwnerTypeUser, "u", now))
		require.NoError(t, v.Revoke("fraud", now))
		require.Equal(t, voucher.StatusRevoked, v.Status)
	})
}

func TestVoucher_Revoke_AfterRedeemed_IsRejected(t *testing.T) {
	v := newTestVoucher(t)
	now := time.Now()
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, nil, now))
	require.NoError(t, v.Activate(voucher.OwnerTypeUser, "u", now))
	require.NoError(t, v.Redeem(shared.NewMoney(100000, "VND"), "", now))

	err := v.Revoke("too late", now)
	require.ErrorIs(t, err, voucher.ErrInvalidTransition)
}

func TestVoucher_Expire_FromIssuedAndActive_Succeeds(t *testing.T) {
	now := time.Now()

	v := newTestVoucher(t)
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, nil, now))
	require.NoError(t, v.Expire(now))
	require.Equal(t, voucher.StatusExpired, v.Status)

	v2 := newTestVoucher(t)
	require.NoError(t, v2.Issue(shared.VoucherCode{Code: "C2"}, nil, now))
	require.NoError(t, v2.Activate(voucher.OwnerTypeUser, "u", now))
	require.NoError(t, v2.Expire(now))
	require.Equal(t, voucher.StatusExpired, v2.Status)
}

func TestVoucher_PullEvents_DrainsAndClears(t *testing.T) {
	v := newTestVoucher(t)
	now := time.Now()
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, nil, now))

	events := v.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, voucher.EventTypeVoucherIssued, events[0].EventType())

	require.Empty(t, v.PullEvents())
}

func TestVoucher_CanRedeem(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)

	v := newTestVoucher(t)
	require.False(t, v.CanRedeem(now), "created voucher should not be redeemable")

	require.NoError(t, v.Issue(shared.VoucherCode{Code: "C"}, &future, now))
	require.False(t, v.CanRedeem(now), "issued (not yet active) voucher should not be redeemable")

	require.NoError(t, v.Activate(voucher.OwnerTypeUser, "u", now))
	require.True(t, v.CanRedeem(now))
}
