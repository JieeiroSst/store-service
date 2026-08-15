package consumer

import (
	"context"
	"encoding/json"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"go.uber.org/zap"
)

const paymentSettledTopic = "payment.events"
const paymentSettledGroup = "voucher-service-order-fulfillment"

type PaymentSettledConsumer struct {
	reader   *Reader
	orderSvc orderapp.OrderService
	log      *zap.Logger
}

func NewPaymentSettledConsumer(factory ReaderFactory, orderSvc orderapp.OrderService, log *zap.Logger) *PaymentSettledConsumer {
	return &PaymentSettledConsumer{
		reader:   factory(paymentSettledTopic, paymentSettledGroup),
		orderSvc: orderSvc,
		log:      log,
	}
}

func (c *PaymentSettledConsumer) Run(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.log.Error("payment settled consumer: read failed", zap.Error(err))
			continue
		}

		var evt paymentapp.PaymentSettledEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			c.log.Error("payment settled consumer: bad payload", zap.Error(err))
			continue
		}
		orderID, err := shared.ParseOrderID(evt.OrderID)
		if err != nil {
			c.log.Error("payment settled consumer: bad order id", zap.Error(err))
			continue
		}
		if err := c.orderSvc.ConfirmPayment(ctx, orderID, evt.PaymentRef); err != nil {
			c.log.Error("payment settled consumer: confirm payment failed", zap.String("order_id", evt.OrderID), zap.Error(err))
		}
	}
}

func (c *PaymentSettledConsumer) Close() error {
	return c.reader.Close()
}
