package channel

import (
	"context"
	"fmt"
	"sync"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
)

type Message struct {
	Recipient string
	Subject   string
	Body      string
	Data      map[string]string 
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
	Type() model.ChannelType
}

type Registry struct {
	mu      sync.RWMutex
	senders map[model.ChannelType]Sender
}

func NewRegistry() *Registry {
	return &Registry{senders: make(map[model.ChannelType]Sender)}
}

func (r *Registry) Register(s Sender) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.senders[s.Type()] = s
}

func (r *Registry) Get(t model.ChannelType) (Sender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.senders[t]
	if !ok {
		return nil, fmt.Errorf("channel %q not registered", t)
	}
	return s, nil
}

func (r *Registry) Send(ctx context.Context, channelType model.ChannelType, msgs []Message) []error {
	s, err := r.Get(channelType)
	if err != nil {
		return []error{err}
	}
	errs := make([]error, 0, len(msgs))
	for _, m := range msgs {
		if err := s.Send(ctx, m); err != nil {
			errs = append(errs, fmt.Errorf("recipient %s: %w", m.Recipient, err))
		}
	}
	return errs
}
