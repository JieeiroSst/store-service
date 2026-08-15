package shared_test

import (
	"testing"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestMoney_Add(t *testing.T) {
	a := shared.NewMoney(1000, "VND")
	b := shared.NewMoney(500, "VND")
	sum, err := a.Add(b)
	require.NoError(t, err)
	require.Equal(t, int64(1500), sum.Amount)
}

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	a := shared.NewMoney(1000, "VND")
	b := shared.NewMoney(500, "USD")
	_, err := a.Add(b)
	require.ErrorIs(t, err, shared.ErrCurrencyMismatch)
}

func TestMoney_Sub(t *testing.T) {
	a := shared.NewMoney(1000, "VND")
	b := shared.NewMoney(300, "VND")
	diff, err := a.Sub(b)
	require.NoError(t, err)
	require.Equal(t, int64(700), diff.Amount)
}

func TestMoney_LessThan(t *testing.T) {
	require.True(t, shared.NewMoney(100, "VND").LessThan(shared.NewMoney(200, "VND")))
	require.False(t, shared.NewMoney(200, "VND").LessThan(shared.NewMoney(100, "VND")))
}

func TestMoney_GreaterThanOrEqual(t *testing.T) {
	require.True(t, shared.NewMoney(200, "VND").GreaterThanOrEqual(shared.NewMoney(200, "VND")))
	require.True(t, shared.NewMoney(300, "VND").GreaterThanOrEqual(shared.NewMoney(200, "VND")))
	require.False(t, shared.NewMoney(100, "VND").GreaterThanOrEqual(shared.NewMoney(200, "VND")))
}

func TestMoney_IsZero(t *testing.T) {
	require.True(t, shared.ZeroMoney("VND").IsZero())
	require.False(t, shared.NewMoney(1, "VND").IsZero())
}
