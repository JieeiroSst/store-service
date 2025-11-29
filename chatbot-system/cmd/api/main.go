package main

import (
	"database/sql"
	"log"
	"net/http"

	"chatbot-system/internal/application"
	"chatbot-system/internal/config"
	"chatbot-system/internal/infrastructure/ai"
	"chatbot-system/internal/infrastructure/database"
	httpHandler "chatbot-system/internal/infrastructure/http"
	"chatbot-system/internal/infrastructure/websocket"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connected successfully")

	// Initialize repositories
	userRepo := database.NewMySQLUserRepository(db)
	conversationRepo := database.NewMySQLConversationRepository(db)
	messageRepo := database.NewMySQLMessageRepository(db)

	// Initialize AI provider factory
	aiFactory := ai.NewProviderFactory(cfg.AI.ClaudeAPIKey, cfg.AI.DeepSeekAPIKey)
	log.Printf("Available AI providers: %v", aiFactory.ListAvailableProviders())

	// Initialize use cases
	chatUseCase := application.NewChatUseCase(userRepo, conversationRepo, messageRepo, aiFactory)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	wsHandler := websocket.NewHandler(hub, chatUseCase)
	chatHandler := httpHandler.NewChatHandler(chatUseCase)

	// Setup router
	router := mux.NewRouter()

	// Register HTTP routes
	chatHandler.RegisterRoutes(router)

	// Register WebSocket route
	router.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure properly in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
