package handlers

import (
	"time"

	"github.com/JIeeiroSst/room-service/internal/core/ports"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine      *gin.Engine
	roomService ports.RoomService
	authService ports.AuthService
	wsHandler   *WebSocketHandler
}

func (r *Router) Run(port string) error {
	r.engine.Run(port)
	return nil
}

func NewRouter(
	roomService ports.RoomService,
	authService ports.AuthService,
	wsHandler *WebSocketHandler,
) *Router {
	router := &Router{
		engine:      gin.Default(),
		roomService: roomService,
		authService: authService,
		wsHandler:   wsHandler,
	}
	router.setupRoutes()
	return router
}

func (r *Router) setupRoutes() {
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.engine.POST("/rooms", r.createRoom)
	r.engine.GET("/rooms", r.listRooms)
	r.engine.POST("/rooms/:id/join", r.joinRoom)

	authenticated := r.engine.Group("/")
	authenticated.Use(r.authMiddleware())
	{
		authenticated.GET("/rooms/:id/messages", r.getMessages)
		authenticated.GET("/ws", r.wsHandler.HandleWebSocket)
	}
}
