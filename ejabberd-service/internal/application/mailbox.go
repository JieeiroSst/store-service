package application

import (
	"sync"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type mailboxItem struct {
	client    *domain.Element
	event     *port.SessionEvent
	terminate *string
	tick      bool
}

type mailbox struct {
	mu     sync.Mutex
	items  []mailboxItem
	signal chan struct{}
	closed bool
}

func newMailbox() *mailbox {
	return &mailbox{signal: make(chan struct{}, 1)}
}

func (m *mailbox) push(it mailboxItem) bool {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return false
	}
	m.items = append(m.items, it)
	m.mu.Unlock()
	select {
	case m.signal <- struct{}{}:
	default:
	}
	return true
}

func (m *mailbox) take() []mailboxItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := m.items
	m.items = nil
	return items
}

func (m *mailbox) close() []mailboxItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	items := m.items
	m.items = nil
	return items
}
