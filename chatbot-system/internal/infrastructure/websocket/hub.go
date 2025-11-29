package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Client struct {
	ID     int64
	Hub    *Hub
	Conn   Connection
	Send   chan []byte
	UserID int64
}

type Connection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

type Hub struct {
	clients    map[int64]*Client
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	RecipientID int64
	Data        []byte
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client registered: UserID=%d", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: UserID=%d", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			client, ok := h.clients[message.RecipientID]
			h.mu.RUnlock()

			if ok {
				select {
				case client.Send <- message.Data:
				default:
					h.mu.Lock()
					close(client.Send)
					delete(h.clients, client.UserID)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *Hub) SendToUser(userID int64, data []byte) {
	h.broadcast <- &BroadcastMessage{
		RecipientID: userID,
		Data:        data,
	}
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (c *Client) ReadPump(messageHandler func([]byte) error) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error for UserID=%d: %v", c.UserID, err)
			break
		}

		if err := messageHandler(message); err != nil {
			log.Printf("Message handler error for UserID=%d: %v", c.UserID, err)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(1, message); err != nil {
			log.Printf("Write error for UserID=%d: %v", c.UserID, err)
			return
		}
	}
}
