package handlers

import (
	"net/http"

	websocketHub "github.com/JIeeiroSst/room-service/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub      *websocketHub.Hub
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(hub *websocketHub.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	roomID := c.GetUint("room_id")
	username := c.GetString("username")

	client := websocketHub.NewClient(h.hub, conn, roomID, username)
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
