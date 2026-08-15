package order_test

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func newTestOrder(t *testing.T) *order.Order {
	t.Helper()
	o, err := order.NewOrder(order.BuyerTypeRetail, "buyer-1", "VND", "idem-key", time.Now())
	require.NoError(t, err)
	require.Equal(t, order.StatusPending, o.Status)
	return o
}

func TestOrder_AddItem_AccumulatesTotal(t *testing.T) {
	o := newTestOrder(t)
	item1, err := order.NewOrderItem(shared.NewMerchantID(), "sku-1", 2, shared.NewMoney(10000, "VND"))
	require.NoError(t, err)
	item2, err := order.NewOrderItem(shared.NewMerchantID(), "sku-2", 1, shared.NewMoney(5000, "VND"))
	require.NoError(t, err)

	require.NoError(t, o.AddItem(item1, time.Now()))
	require.NoError(t, o.AddItem(item2, time.Now()))

	require.Equal(t, int64(25000), o.TotalAmount.Amount)
}

func TestOrder_MarkAwaitingPayment_RequiresItems(t *testing.T) {
	o := newTestOrder(t)
	err := o.MarkAwaitingPayment(time.Now())
	require.ErrorIs(t, err, order.ErrEmptyOrder)
}

func TestOrder_FullLifecycle_Succeeds(t *testing.T) {
	o := newTestOrder(t)
	item, _ := order.NewOrderItem(shared.NewMerchantID(), "sku-1", 1, shared.NewMoney(10000, "VND"))
	require.NoError(t, o.AddItem(item, time.Now()))

	require.NoError(t, o.MarkAwaitingPayment(time.Now()))
	require.Equal(t, order.StatusAwaitingPayment, o.Status)

	require.NoError(t, o.MarkPaid("txn-1", time.Now()))
	require.Equal(t, order.StatusPaid, o.Status)

	require.NoError(t, o.MarkFulfilling(time.Now()))
	require.Equal(t, order.StatusFulfilling, o.Status)

	require.NoError(t, o.Complete(time.Now()))
	require.Equal(t, order.StatusCompleted, o.Status)
}

func TestOrder_Cancel_AfterCompleted_IsRejected(t *testing.T) {
	o := newTestOrder(t)
	item, _ := order.NewOrderItem(shared.NewMerchantID(), "sku-1", 1, shared.NewMoney(10000, "VND"))
	require.NoError(t, o.AddItem(item, time.Now()))
	require.NoError(t, o.MarkAwaitingPayment(time.Now()))
	require.NoError(t, o.MarkPaid("txn-1", time.Now()))
	require.NoError(t, o.MarkFulfilling(time.Now()))
	require.NoError(t, o.Complete(time.Now()))

	err := o.Cancel("changed mind", time.Now())
	require.ErrorIs(t, err, order.ErrInvalidOrderTransition)
}

func TestOrder_Cancel_FromPending_Succeeds(t *testing.T) {
	o := newTestOrder(t)
	require.NoError(t, o.Cancel("changed mind", time.Now()))
	require.Equal(t, order.StatusCancelled, o.Status)
}

func TestOrder_AddItem_AfterAwaitingPayment_IsRejected(t *testing.T) {
	o := newTestOrder(t)
	item, _ := order.NewOrderItem(shared.NewMerchantID(), "sku-1", 1, shared.NewMoney(10000, "VND"))
	require.NoError(t, o.AddItem(item, time.Now()))
	require.NoError(t, o.MarkAwaitingPayment(time.Now()))

	item2, _ := order.NewOrderItem(shared.NewMerchantID(), "sku-2", 1, shared.NewMoney(1000, "VND"))
	err := o.AddItem(item2, time.Now())
	require.ErrorIs(t, err, order.ErrInvalidOrderTransition)
}
