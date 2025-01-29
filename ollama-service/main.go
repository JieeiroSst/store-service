package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatMessage represents a single message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// LlamaRequest represents the request to the Llama service
type LlamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

// StreamResponse represents the streaming response from Llama
type StreamResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// ChatResponse represents the API response
type ChatResponse struct {
	Response string `json:"response"`
}

func main() {
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Chat endpoint
	r.POST("/chat", handleChat)

	// Start server
	log.Fatal(r.Run(":8082"))
}

func handleChat(c *gin.Context) {
	var chatReq ChatRequest
	if err := c.BindJSON(&chatReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create Llama request
	llamaReq := LlamaRequest{
		Model:    "llama3.2-vision",
		Messages: chatReq.Messages,
		Stream:   true,
	}

	// Send request to Llama service
	response, err := sendToLlama(llamaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: response,
	})
}

func sendToLlama(req LlamaRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var fullResponse string

	for {
		var streamResp StreamResponse
		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		fullResponse += streamResp.Message.Content

		if streamResp.Done {
			break
		}
	}

	return fullResponse, nil
}