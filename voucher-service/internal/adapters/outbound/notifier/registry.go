package notifier

import (
	"fmt"

	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
)

type Registry struct {
	byChannel map[notificationapp.Channel]notificationapp.Notifier
}

func NewRegistry(notifiers []notificationapp.Notifier) notificationapp.NotifierRegistry {
	byChannel := make(map[notificationapp.Channel]notificationapp.Notifier, len(notifiers))
	for _, n := range notifiers {
		byChannel[n.Channel()] = n
	}
	return &Registry{byChannel: byChannel}
}

func (r *Registry) Resolve(channel notificationapp.Channel) (notificationapp.Notifier, error) {
	n, ok := r.byChannel[channel]
	if !ok {
		return nil, fmt.Errorf("no notifier registered for channel %q", channel)
	}
	return n, nil
}
