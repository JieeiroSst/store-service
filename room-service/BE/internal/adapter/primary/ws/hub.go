package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[uint]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: map[uint]map[*Client]struct{}{}}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.roomID] == nil {
		h.rooms[c.roomID] = map[*Client]struct{}{}
	}
	h.rooms[c.roomID][c] = struct{}{}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.rooms[c.roomID]; clients != nil {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, c.roomID)
		}
	}
}

func (h *Hub) Deliver(e model.Event) {
	frame, err := json.Marshal(e)
	if err != nil {
		log.Printf("ws: marshal event: %v", err)
		return
	}

	h.mu.RLock()
	var slow, kicked []*Client
	for c := range h.rooms[e.RoomID] {
		if !c.enqueue(frame) {
			slow = append(slow, c)
		} else if e.Type == model.EventMemberRemoved && e.Member != nil && e.Member.Username == c.user.Username {
			kicked = append(kicked, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range slow {
		c.close()
	}
	for _, c := range kicked {
		c.closeAfterFlush()
	}
}

func (h *Hub) Count(roomID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}
