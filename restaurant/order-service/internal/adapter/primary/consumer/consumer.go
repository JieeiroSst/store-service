package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	usecase port.OrderUsecase
	nc      *nats.Conn
	subs    []*nats.Subscription
}

func NewSubscriber(usecase port.OrderUsecase, nc *nats.Conn) *Subscriber {
	return &Subscriber{usecase: usecase, nc: nc}
}

func (s *Subscriber) Start() error {
	subs := []struct {
		subject string
		handler nats.MsgHandler
	}{
		{"order.created", s.onOrderCreated},
		{"order.reject", s.onOrderReject},
		{"order.success", s.onOrderSuccess},
	}

	for _, sub := range subs {
		nsub, err := s.nc.Subscribe(sub.subject, sub.handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, nsub)
	}
	return nil
}

func (s *Subscriber) Stop() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

func (s *Subscriber) onOrderCreated(msg *nats.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
		return
	}

	if err := s.usecase.CreateOrder(context.Background(), &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
	}
}

func (s *Subscriber) onOrderReject(msg *nats.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
		return
	}

	if err := s.usecase.CancelOrder(context.Background(), order.ID, &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
	}
}

func (s *Subscriber) onOrderSuccess(msg *nats.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
		return
	}

	if err := s.usecase.SuccessOrder(context.Background(), order.ID, &order); err != nil {
		logger.ConfigZap().Errorf("%v", err)
	}
}
