package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chatbot-system/internal/application"

	"github.com/gorilla/mux"
)

type ChatHandler struct {
	chatUseCase *application.ChatUseCase
}

func NewChatHandler(chatUseCase *application.ChatUseCase) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
	}
}

type GetHistoryRequest struct {
	UserID         int64 `json:"user_id"`
	ConversationID int64 `json:"conversation_id"`
	Limit          int   `json:"limit"`
	Offset         int   `json:"offset"`
}

func (h *ChatHandler) GetConversationHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["conversation_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// In production, get userID from authentication
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	messages, err := h.chatUseCase.GetConversationHistory(r.Context(), userID, conversationID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *ChatHandler) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	// In production, get userID from authentication
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	conversations, err := h.chatUseCase.GetUserConversations(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversations": conversations,
		"count":         len(conversations),
	})
}

func (h *ChatHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/conversations/{conversation_id}/history", h.GetConversationHistory).Methods("GET")
	router.HandleFunc("/api/conversations", h.GetUserConversations).Methods("GET")
}
