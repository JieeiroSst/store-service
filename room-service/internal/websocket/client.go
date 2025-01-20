package websocket

import (
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	room     uint
	username string
	send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, room uint, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		room:     room,
		username: username,
		send:     make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msg.RoomID = c.room
		msg.Username = c.username
		msg.Timestamp = time.Now()

		c.hub.Mutex.RLock()
		for client := range c.hub.Rooms[c.room] {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(c.hub.Rooms[c.room], client)
			}
		}
		c.hub.Mutex.RUnlock()
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
