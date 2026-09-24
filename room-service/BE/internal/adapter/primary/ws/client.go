package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingInterval = pongWait * 9 / 10
	maxFrameSize = 16 << 10
	sendBuffer   = 64
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	user   model.User
	roomID uint
	chat   port.ChatService

	send      chan []byte
	closeOnce sync.Once
	done      chan struct{}
	flush     chan struct{}
	flushOnce sync.Once
}

func (h *Hub) Serve(conn *websocket.Conn, user model.User, roomID uint, chat port.ChatService) {
	c := &Client{
		hub: h, conn: conn, user: user, roomID: roomID, chat: chat,
		send:  make(chan []byte, sendBuffer),
		done:  make(chan struct{}),
		flush: make(chan struct{}),
	}
	h.register(c)
	go c.writePump()
	c.readPump()
}

func (c *Client) enqueue(frame []byte) bool {
	select {
	case <-c.done:
		return true
	default:
	}
	select {
	case c.send <- frame:
		return true
	default:
		return false
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.hub.unregister(c)
		c.conn.Close()
	})
}

func (c *Client) closeAfterFlush() {
	c.flushOnce.Do(func() { close(c.flush) })
}

func (c *Client) readPump() {
	defer c.close()
	c.conn.SetReadLimit(maxFrameSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var in struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(raw, &in); err != nil {
			c.reply("invalid frame")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = c.chat.Send(ctx, c.user, c.roomID, in.Content)
		cancel()
		switch {
		case err == nil:
		case errors.Is(err, port.ErrInvalidInput):
			c.reply("message is empty or too long")
		case errors.Is(err, port.ErrForbidden), errors.Is(err, port.ErrNotFound):
			return
		default:
			c.reply("could not send message")
		}
	}
}

func (c *Client) reply(msg string) {
	frame, _ := json.Marshal(map[string]string{"type": "error", "error": msg})
	c.enqueue(frame)
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	write := func(msgType int, data []byte) error {
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		return c.conn.WriteMessage(msgType, data)
	}

	for {
		select {
		case frame := <-c.send:
			if write(websocket.TextMessage, frame) != nil {
				return
			}
		case <-c.flush:
			for {
				select {
				case frame := <-c.send:
					if write(websocket.TextMessage, frame) != nil {
						return
					}
				default:
					_ = write(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "removed from room"))
					return
				}
			}
		case <-ticker.C:
			if write(websocket.PingMessage, nil) != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
