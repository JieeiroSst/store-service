package http

import (
	"sync"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/gorilla/websocket"
)

type conn struct {
	id     string
	roomID string
	ws     *websocket.Conn
	cfg    config.WebSocketConfig

	mu    sync.Mutex
	media domain.MediaState

	send      chan domain.Message
	done      chan struct{}
	closeOnce sync.Once
}

func newConn(ws *websocket.Conn, userID, roomID string, cfg config.WebSocketConfig) *conn {
	return &conn{
		id:     userID,
		roomID: roomID,
		ws:     ws,
		cfg:    cfg,
		media:  domain.DefaultMediaState(),
		send:   make(chan domain.Message, cfg.SendBuffer),
		done:   make(chan struct{}),
	}
}

func (c *conn) ID() string     { return c.id }
func (c *conn) RoomID() string { return c.roomID }

func (c *conn) Media() domain.MediaState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.media
}

func (c *conn) SetMedia(m domain.MediaState) {
	c.mu.Lock()
	c.media = m
	c.mu.Unlock()
}

func (c *conn) Send(msg domain.Message) error {
	select {
	case <-c.done:
		return domain.ErrPeerClosed
	default:
	}
	select {
	case c.send <- msg:
		return nil
	default:
		return domain.ErrPeerBusy
	}
}

func (c *conn) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}

func (c *conn) writePump(shutdown <-chan struct{}) {
	ticker := time.NewTicker(c.cfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(c.cfg.WriteWait))
			if err := c.ws.WriteJSON(msg); err != nil {
				c.Close()
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(c.cfg.WriteWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Close()
				return
			}
		case <-c.done:

			c.flush()
			c.writeClose(websocket.CloseNormalClosure)
			return
		case <-shutdown:
			c.Close()
			c.writeClose(websocket.CloseGoingAway)
			return
		}
	}
}

func (c *conn) flush() {
	for {
		select {
		case msg := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(c.cfg.WriteWait))
			if err := c.ws.WriteJSON(msg); err != nil {
				return
			}
		default:
			return
		}
	}
}

func (c *conn) writeClose(code int) {
	c.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""), time.Now().Add(c.cfg.WriteWait))
}
