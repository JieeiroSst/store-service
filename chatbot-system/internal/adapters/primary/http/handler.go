package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"chatbot-system/internal/core/domain"
	"chatbot-system/internal/core/ports"
)

type HTTPHandler struct {
	chatService ports.ChatService
}

func NewHTTPHandler(chatService ports.ChatService) *HTTPHandler {
	return &HTTPHandler{
		chatService: chatService,
	}
}

// CreateConversation creates a new conversation
func (h *HTTPHandler) CreateConversation(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		AIModel string `json:"ai_model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.chatService.CreateConversation(c.Request.Context(), req.UserID, req.Title, req.AIModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

// GetChatHistory retrieves chat history for a conversation
func (h *HTTPHandler) GetChatHistory(c *gin.Context) {
	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	messages, err := h.chatService.GetChatHistory(c.Request.Context(), uint(conversationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// SendMessage sends a message (REST endpoint as alternative to WebSocket)
func (h *HTTPHandler) SendMessage(c *gin.Context) {
	var req domain.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatService.ProcessMessage(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SwitchAIModel switches the AI model for a conversation
func (h *HTTPHandler) SwitchAIModel(c *gin.Context) {
	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req struct {
		AIModel string `json:"ai_model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.chatService.SwitchAIModel(c.Request.Context(), uint(conversationID), req.AIModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI model switched successfully"})
}

// HealthCheck endpoint
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
