package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	wsPingInterval = 30 * time.Second
	wsPongWait     = 60 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	chat  port.ChatUsecase
	authn middleware.Authenticator
}

func NewWSHandler(chat port.ChatUsecase, users *userservice.Client) *WSHandler {
	return &WSHandler{chat: chat, authn: users}
}

func chatToken(c *gin.Context) string {
	if t, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return strings.TrimSpace(c.Query("access_token"))
}

type inboundChatMessage struct {
	Body string `json:"body"`
}

func (h *WSHandler) ChatRoom(c *gin.Context) {
	roomID := c.Param("id")

	var userID, username string
	if token := chatToken(c); token != "" {
		id, err := h.authn.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		userID, username = id.UserID, id.Username
		if username == "" {
			username = id.UserID
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	metrics.WebsocketConnections.Inc()
	defer metrics.WebsocketConnections.Dec()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	messages, unsub, err := h.chat.Subscribe(ctx, roomID)
	if err != nil {
		logger.WithContext(ctx).Warn("chat subscribe failed", zap.String("roomId", roomID), zap.Error(err))
		return
	}
	defer unsub()

	localErrors := make(chan string, 4)

	go func() {
		ticker := time.NewTicker(wsPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-messages:
				if !ok {
					cancel()
					return
				}
				if err := conn.WriteJSON(msg); err != nil {
					cancel()
					return
				}
			case errMsg := <-localErrors:
				if err := conn.WriteJSON(gin.H{"error": errMsg}); err != nil {
					cancel()
					return
				}
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	for {
		var in inboundChatMessage
		if err := conn.ReadJSON(&in); err != nil {
			return
		}
		if userID == "" {
			select {
			case localErrors <- "sign in to chat":
			default:
			}
			continue
		}
		msg := model.ChatMessage{
			RoomID:   roomID,
			UserID:   userID,
			Username: username,
			Body:     in.Body,
			SentAt:   time.Now(),
		}
		if err := h.chat.Publish(ctx, msg); err != nil {
			if errors.Is(err, port.ErrBanned) {
				select {
				case localErrors <- "you are banned from this room's chat":
				default:
				}
				continue
			}
			logger.WithContext(ctx).Warn("chat publish failed", zap.String("roomId", roomID), zap.Error(err))
		}
	}
}
