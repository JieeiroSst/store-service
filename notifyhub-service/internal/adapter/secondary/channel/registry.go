package channel

import (
	"fmt"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

type Registry struct {
	senders map[model.ChannelType]port.Sender
}

func NewRegistry(senders ...port.Sender) *Registry {
	r := &Registry{senders: make(map[model.ChannelType]port.Sender, len(senders))}
	for _, s := range senders {
		r.senders[s.Type()] = s
	}
	return r
}

func (r *Registry) Get(t model.ChannelType) (port.Sender, error) {
	s, ok := r.senders[t]
	if !ok {
		return nil, fmt.Errorf("channel %q not configured", t)
	}
	return s, nil
}
