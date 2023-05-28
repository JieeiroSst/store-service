package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/gorilla/websocket"
)

type WebSocket struct {
	Usecase *usecase.Usecase
}

func NewWebSocket(Usecase *usecase.Usecase) *WebSocket {
	return &WebSocket{
		Usecase: Usecase,
	}
}

func (u *WebSocket) SetupRoutes() {
	http.HandleFunc("/save-message", u.WebSocketSaveMessage)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (u *WebSocket) reader(conn *websocket.Conn) {
	var (
		message dto.Messages
	)
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if err := json.Unmarshal(p, &message); err != nil {
			return
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			return
		}

	}
}

func (u *WebSocket) WebSocketSaveMessage(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	err = ws.WriteMessage(1, []byte("Save Message"))
	if err != nil {
		log.Println(err)
	}

	u.reader(ws)
}
