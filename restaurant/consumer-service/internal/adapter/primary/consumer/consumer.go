package consumer

import "github.com/nats-io/nats.go"

// Subscriber currently subscribes to nothing — preserved as an empty
// extension point, matching the original delivery/consumer stub.
type Subscriber struct {
	nc *nats.Conn
}

func NewSubscriber(nc *nats.Conn) *Subscriber {
	return &Subscriber{nc: nc}
}

func (s *Subscriber) Start() error {
	return nil
}

func (s *Subscriber) Stop() {}
