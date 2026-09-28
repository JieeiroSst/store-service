package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chatbot-system/internal/application"
	"chatbot-system/internal/infrastructure/auth"

	"github.com/gorilla/mux"
)

type ChatHandler struct {
	chatUseCase *application.ChatUseCase
	authn       *auth.Authenticator
}

func NewChatHandler(chatUseCase *application.ChatUseCase, authn *auth.Authenticator) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
		authn:       authn,
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

	userID := auth.UserID(r.Context())

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
	userID := auth.UserID(r.Context())

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

// RegisterRoutes mounts the API; every call acts as the user-service
// account behind its bearer token.
func (h *ChatHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api").Subrouter()
	api.Use(h.authn.Middleware)
	api.HandleFunc("/conversations/{conversation_id}/history", h.GetConversationHistory).Methods("GET")
	api.HandleFunc("/conversations", h.GetUserConversations).Methods("GET")
}
