// internal/handlers/router.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"github.com/gin-gonic/gin"
)

// createRoom handles the creation of a new chat room
func (r *Router) createRoom(c *gin.Context) {
	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if room.Name == "" || room.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Room name and password are required",
		})
		return
	}

	if err := r.roomService.CreateRoom(&room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create room",
			"details": err.Error(),
		})
		return
	}

	// Don't return the hashed password
	room.Password = ""
	c.JSON(http.StatusCreated, room)
}

// listRooms returns a list of all available chat rooms
func (r *Router) listRooms(c *gin.Context) {
	rooms, err := r.roomService.ListRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch rooms",
			"details": err.Error(),
		})
		return
	}

	// Don't return passwords in the response
	for i := range rooms {
		rooms[i].Password = ""
	}

	c.JSON(http.StatusOK, rooms)
}

// joinRoom handles user authentication for joining a specific room
func (r *Router) joinRoom(c *gin.Context) {
	var input struct {
		Password string `json:"password" binding:"required"`
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	roomID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	// Validate username length and characters
	if len(input.Username) < 3 || len(input.Username) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username must be between 3 and 20 characters",
		})
		return
	}

	token, err := r.roomService.JoinRoom(uint(roomID), input.Password, input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid room ID or password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// getMessages returns all messages for a specific room
func (r *Router) getMessages(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	// Verify that the user has access to this room
	tokenRoomID, exists := c.Get("room_id")
	if !exists || tokenRoomID.(uint) != uint(roomID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this room",
		})
		return
	}

	messages, err := r.roomService.GetMessages(uint(roomID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch messages",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// authMiddleware validates the JWT token and sets user information in the context
func (r *Router) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			c.Abort()
			return
		}

		claims, err := r.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("room_id", claims.RoomID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// rateLimitMiddleware implements a simple rate limiting mechanism
func (r *Router) rateLimitMiddleware() gin.HandlerFunc {
	// TODO: Implement rate limiting if needed
	return func(c *gin.Context) {
		c.Next()
	}
}

// errorResponse standardizes error responses
type errorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// successResponse standardizes success responses
type successResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
