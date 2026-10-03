package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/JIeeiroSst/webrtc-service/internal/port"
	"github.com/JIeeiroSst/webrtc-service/web"
	"github.com/gorilla/websocket"
)

type Handler struct {
	svc      port.SignalingService
	ice      port.ICEService
	cfg      config.WebSocketConfig
	upgrader websocket.Upgrader

	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func NewHandler(svc port.SignalingService, ice port.ICEService, cfg *config.Config) *Handler {
	return &Handler{
		svc: svc,
		ice: ice,
		cfg: cfg.WebSocket,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     checkOrigin(cfg.Server.AllowedOrigins),
		},
		shutdown: make(chan struct{}),
	}
}

func checkOrigin(allowed []string) func(*http.Request) bool {
	if len(allowed) == 0 {
		return nil
	}
	if slices.Contains(allowed, "*") {
		return func(*http.Request) bool { return true }
	}
	return func(r *http.Request) bool {
		return slices.Contains(allowed, r.Header.Get("Origin"))
	}
}

func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(web.Index)
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ICEServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{
		"ice_servers": h.ice.Servers(r.URL.Query().Get("user_id")),
	})
}

func (h *Handler) Room(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.Room(r.PathValue("room_id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidID) {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, domain.ErrorPayload{Code: domain.ErrorCode(err), Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	roomID := r.URL.Query().Get("room_id")
	if domain.ValidateID(userID) != nil || domain.ValidateID(roomID) != nil {
		http.Error(w, domain.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {

		log.Printf("websocket upgrade: %v", err)
		return
	}

	c := newConn(ws, userID, roomID, h.cfg)
	go c.writePump(h.shutdown)

	if err := h.svc.Join(c); err != nil {
		c.Send(domain.NewErrorMessage(err))
		c.Close()
		return
	}
	defer func() {
		h.svc.Leave(c)
		c.Close()
	}()
	h.readLoop(c)
}

func (h *Handler) readLoop(c *conn) {
	c.ws.SetReadLimit(h.cfg.MaxMessageBytes)
	c.ws.SetReadDeadline(time.Now().Add(h.cfg.PongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(h.cfg.PongWait))
	})

	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Printf("websocket read (user %s, room %s): %v", c.id, c.roomID, err)
			}
			return
		}

		var msg domain.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.Send(domain.NewErrorMessage(domain.ErrInvalidMessage))
			continue
		}
		if err := h.svc.Relay(c, msg); err != nil {
			c.Send(domain.NewErrorMessage(err))
		}
	}
}

func (h *Handler) Shutdown() {
	h.shutdownOnce.Do(func() { close(h.shutdown) })
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
