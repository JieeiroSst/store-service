package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"github.com/nats-io/nats.go"
)

const cartAuthorizeSubject = "accounting.authorize"

type Subscriber struct {
	usecase port.AuthCartUsecase
	nc      *nats.Conn
	subs    []*nats.Subscription
}

func NewSubscriber(usecase port.AuthCartUsecase, nc *nats.Conn) *Subscriber {
	return &Subscriber{usecase: usecase, nc: nc}
}

func (s *Subscriber) Start() error {
	sub, err := s.nc.Subscribe(cartAuthorizeSubject, s.onCartAuthorize)
	if err != nil {
		return err
	}
	s.subs = append(s.subs, sub)
	return nil
}

func (s *Subscriber) Stop() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

func (s *Subscriber) onCartAuthorize(msg *nats.Msg) {
	var cart model.AuthCart
	if err := json.Unmarshal(msg.Data, &cart); err != nil {
		logger.ConfigZap().Error(err)
		return
	}

	if err := s.usecase.PlaceOrder(context.Background(), cart); err != nil {
		logger.ConfigZap().Error(err)
	}
}
