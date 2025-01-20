package websocket

import (
	"sync"
)

type Hub struct {
	Rooms      map[uint]map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[uint]map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			if _, ok := h.Rooms[client.room]; !ok {
				h.Rooms[client.room] = make(map[*Client]bool)
			}
			h.Rooms[client.room][client] = true
			h.Mutex.Unlock()

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Rooms[client.room]; ok {
				if _, ok := h.Rooms[client.room][client]; ok {
					delete(h.Rooms[client.room], client)
					close(client.send)
				}
			}
			h.Mutex.Unlock()
		}
	}
}
