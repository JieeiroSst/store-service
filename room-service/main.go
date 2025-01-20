package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	RoomID    uint      `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "text" or "image"
	Timestamp time.Time `json:"timestamp"`
}

type Room struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name"`
	Password    string    `json:"password"`
	ActiveUsers int       `json:"activeUsers" gorm:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Client struct {
	conn     *websocket.Conn
	room     uint
	username string
	send     chan []byte
}

type Hub struct {
	rooms      map[uint]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type TokenClaims struct {
	Username string `json:"username"`
	RoomID   uint   `json:"room_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("JWT_SECRET_KEY")

var (
	db       *gorm.DB
	hub      *Hub
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func generateToken(roomID uint, username string) (string, error) {
	claims := TokenClaims{
		Username: username,
		RoomID:   roomID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		}

		if token == "" {
			c.JSON(401, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims := &TokenClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !parsedToken.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("room_id", claims.RoomID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func initDB() {
	dsn := "chatuser:chatpass123@tcp(localhost:3306)/chatdb?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Room{}, &Message{})
}

func newHub() *Hub {
	return &Hub{
		rooms:      make(map[uint]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if _, ok := h.rooms[client.room]; !ok {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.rooms[client.room]; ok {
				if _, ok := h.rooms[client.room][client]; ok {
					delete(h.rooms[client.room], client)
					close(client.send)
				}
			}
			h.mutex.Unlock()
		}
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	roomID, _ := strconv.ParseUint(c.Query("room_id"), 10, 32)
	username := c.GetString("username")

	client := &Client{
		conn:     conn,
		room:     uint(roomID),
		username: username,
		send:     make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msg.RoomID = c.room
		msg.Username = c.username
		msg.Timestamp = time.Now()

		if !strings.Contains(msg.Type, "typing") {
			db.Create(&msg)
		}

		hub.mutex.RLock()
		for client := range hub.rooms[c.room] {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(hub.rooms[c.room], client)
			}
		}
		hub.mutex.RUnlock()
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func main() {
	initDB()
	hub = newHub()
	go hub.run()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/rooms", createRoom)
	r.GET("/rooms", listRooms)
	r.POST("/rooms/:id/join", joinRoom)

	authenticated := r.Group("/")
	authenticated.Use(authMiddleware())
	{
		authenticated.GET("/rooms/:id/messages", getMessages)
		authenticated.GET("/ws", handleWebSocket)
	}

	r.Run(":8081")
}

func createRoom(c *gin.Context) {
	var room Room
	if err := c.BindJSON(&room); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(room.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	room.Password = string(hashedPassword)

	db.Create(&room)
	c.JSON(200, room)
}

func listRooms(c *gin.Context) {
	var rooms []Room
	result := db.Find(&rooms)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load rooms",
		})
		return
	}

	for i := range rooms {
		rooms[i].ActiveUsers = 0
	}

	c.JSON(http.StatusOK, rooms)
}

func joinRoom(c *gin.Context) {
	var input struct {
		Password string `json:"password"`
		Username string `json:"username"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var room Room
	if err := db.First(&room, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "Room not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(room.Password), []byte(input.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	token, err := generateToken(room.ID, input.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(200, gin.H{"token": token})
}

func getMessages(c *gin.Context) {
	var messages []Message
	roomID := c.Param("id")
	query := db.Where("room_id = ?", roomID)
	query.Order("timestamp asc").Find(&messages)
	c.JSON(200, messages)
}
