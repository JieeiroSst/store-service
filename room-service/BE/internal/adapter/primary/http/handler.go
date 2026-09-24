package http

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/adapter/primary/ws"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const userKey = "user"

type Handler struct {
	auth     port.AuthService
	rooms    port.RoomService
	chat     port.ChatService
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewHandler(auth port.AuthService, rooms port.RoomService, chat port.ChatService, hub *ws.Hub, cfg *config.Config) *Handler {
	return &Handler{
		auth: auth, rooms: rooms, chat: chat, hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     originChecker(cfg.Server.AllowedOrigins),
		},
	}
}

func originChecker(origins []string) func(*http.Request) bool {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || allowed["*"] || allowed[origin]
	}
}

func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, port.ErrInvalidInput)
		return
	}
	session, err := h.auth.Login(c.Request.Context(), in.Username, in.Password)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, port.ErrInvalidInput)
		return
	}
	session, err := h.auth.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			token = c.Query("token")
		}
		user, err := h.auth.Authenticate(c.Request.Context(), token)
		if err != nil {
			fail(c, err)
			c.Abort()
			return
		}
		c.Set(userKey, user)
		c.Next()
	}
}

func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, actor(c))
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var in struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, port.ErrInvalidInput)
		return
	}
	room, err := h.rooms.CreateRoom(c.Request.Context(), actor(c), in.Name)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, room)
}

func (h *Handler) ListRooms(c *gin.Context) {
	rooms, err := h.rooms.ListRooms(c.Request.Context(), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, nonNil(rooms))
}

func (h *Handler) GetRoom(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	room, err := h.rooms.GetRoom(c.Request.Context(), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *Handler) ListMembers(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	members, err := h.rooms.ListMembers(c.Request.Context(), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, nonNil(members))
}

func (h *Handler) AddMember(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	var in struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, port.ErrInvalidInput)
		return
	}
	m, err := h.rooms.AddMember(c.Request.Context(), actor(c), id, in.Username)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	if err := h.rooms.RemoveMember(c.Request.Context(), actor(c), id, c.Param("username")); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Messages(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	before, _ := strconv.ParseUint(c.Query("before"), 10, 32)
	limit, _ := strconv.Atoi(c.Query("limit"))
	msgs, err := h.chat.History(c.Request.Context(), actor(c), id, uint(before), limit)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, nonNil(msgs))
}

// Connect upgrades to a WebSocket for live chat in one room. Membership is
// checked before the upgrade so a non-member gets a plain HTTP error.
func (h *Handler) Connect(c *gin.Context) {
	id, ok := roomID(c)
	if !ok {
		return
	}
	user := actor(c)
	if err := h.chat.Authorize(c.Request.Context(), user, id); err != nil {
		fail(c, err)
		return
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Serve(conn, user, id, h.chat)
}

func actor(c *gin.Context) model.User {
	return c.MustGet(userKey).(model.User)
}

func roomID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		fail(c, port.ErrInvalidInput)
		return 0, false
	}
	return uint(id), true
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func fail(c *gin.Context, err error) {
	status, msg := http.StatusInternalServerError, "internal error"
	switch {
	case errors.Is(err, port.ErrInvalidInput):
		status, msg = http.StatusBadRequest, "invalid input"
	case errors.Is(err, port.ErrInvalidLogin):
		status, msg = http.StatusUnauthorized, err.Error()
	case errors.Is(err, port.ErrUnauthenticated):
		status, msg = http.StatusUnauthorized, "unauthenticated"
	case errors.Is(err, port.ErrForbidden):
		status, msg = http.StatusForbidden, "forbidden"
	case errors.Is(err, port.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, port.ErrAlreadyMember):
		status, msg = http.StatusConflict, "already a member"
	case errors.Is(err, port.ErrUnavailable):
		status, msg = http.StatusBadGateway, "user service unavailable"
	default:
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	}
	c.JSON(status, gin.H{"error": msg})
}
