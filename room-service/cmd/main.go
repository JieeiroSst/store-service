// cmd/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/room-service/internal/config"
	"github.com/JIeeiroSst/room-service/internal/core/ports"
	"github.com/JIeeiroSst/room-service/internal/core/services"
	"github.com/JIeeiroSst/room-service/internal/handlers"
	"github.com/JIeeiroSst/room-service/internal/repositories"
	"github.com/JIeeiroSst/room-service/internal/websocket"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Modules struct {
	fx.In

	DB          *gorm.DB
	RoomService ports.RoomService
	AuthService ports.AuthService
	RoomRepo    ports.RoomRepository
	MessageRepo ports.MessageRepository
	Hub         *websocket.Hub
	Router      *handlers.Router
	WSHandler   *handlers.WebSocketHandler
}

func main() {
	app := fx.New(
		// Provide all dependencies
		fx.Provide(
			// Configuration
			config.NewConfig,
			config.NewDB,

			// Repositories
			repositories.NewRoomRepository,
			repositories.NewMessageRepository,

			// WebSocket
			websocket.NewHub,

			// Services
			services.NewAuthService,
			services.NewRoomService,

			// Handlers
			handlers.NewWebSocketHandler,
			handlers.NewRouter,
		),

		// Invoke functions to start the application
		fx.Invoke(
			registerHooks,
		),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}

// registerHooks sets up the application lifecycle hooks
func registerHooks(lc fx.Lifecycle, m Modules) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			// Start the WebSocket hub
			go m.Hub.Run()

			// Start the HTTP server
			go func() {
				if err := m.Router.Run(":8081"); err != nil {
					log.Printf("Error starting server: %v\n", err)
				}
			}()

			log.Println("Application started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down application...")
			return nil
		},
	})
}

// Bootstrap configuration for development
func init() {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)
}
