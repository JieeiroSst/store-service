package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/JIeeiroSst/ollama-service/internal/userservice"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "chatapp"
)

var db *sql.DB

type Message struct {
	ID          int64  `json:"id,omitempty"`
	Type        string `json:"type"`
	MessageType string `json:"messageType"`
	SenderID    string `json:"sender_id"`
	RecipientID string `json:"recipient_id,omitempty"`
	GroupID     int64  `json:"group_id,omitempty"`
	Data        struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type StreamResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   int64     `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupMember struct {
	GroupID  int64     `json:"group_id"`
	UserID   int64     `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Client struct {
	Conn     *websocket.Conn
	SendChan chan Message
	UserID   int64
	Groups   map[int64]bool 
}

type ChatServer struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type InviteToGroupRequest struct {
	GroupID int64  `json:"group_id" binding:"required"`
	UserID  int64  `json:"user_id" binding:"required"`
	Role    string `json:"role" binding:"required"`
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message, 1000),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
	}
}

func InitDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createTables creates the chat schema. Users live in user-service, so user
// ids are plain BIGINT references (user-service ids) with no local FK.
func createTables(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS groups (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			creator_id BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL,
			role VARCHAR(50) NOT NULL,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			message_type VARCHAR(50) NOT NULL,
			sender_id BIGINT,
			recipient_id BIGINT NULL,
			group_id INTEGER REFERENCES groups(id) NULL,
			content TEXT NOT NULL,
			url TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		// Databases created before users moved to user-service: drop the FKs
		// to the old local users table and widen ids to user-service's range.
		// The old users table itself is left for a manual cleanup.
		`ALTER TABLE groups DROP CONSTRAINT IF EXISTS groups_creator_id_fkey`,
		`ALTER TABLE groups ALTER COLUMN creator_id TYPE BIGINT`,
		`ALTER TABLE group_members DROP CONSTRAINT IF EXISTS group_members_user_id_fkey`,
		`ALTER TABLE group_members ALTER COLUMN user_id TYPE BIGINT`,
		`ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_sender_id_fkey`,
		`ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_recipient_id_fkey`,
		`ALTER TABLE messages ALTER COLUMN sender_id TYPE BIGINT`,
		`ALTER TABLE messages ALTER COLUMN recipient_id TYPE BIGINT`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (server *ChatServer) Run() {
	for {
		select {
		case client := <-server.Register:
			server.Mutex.Lock()
			server.Clients[client] = true
			server.Mutex.Unlock()

			server.loadUserGroups(client)

		case client := <-server.Unregister:
			server.Mutex.Lock()
			if _, ok := server.Clients[client]; ok {
				delete(server.Clients, client)
				close(client.SendChan)
			}
			server.Mutex.Unlock()

		case message := <-server.Broadcast:
			if message.GroupID > 0 {
				server.broadcastToGroup(message)
			} else if message.RecipientID != "" {
				server.sendDirectMessage(message)
			} else {
				server.broadcastToAll(message)
			}
		}
	}
}

func (server *ChatServer) loadUserGroups(client *Client) {
	rows, err := db.Query("SELECT group_id FROM group_members WHERE user_id = $1", client.UserID)
	if err != nil {
		log.Println("Error loading user groups:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			log.Println("Error scanning group ID:", err)
			continue
		}
		client.Groups[groupID] = true
	}
}

func (server *ChatServer) broadcastToGroup(message Message) {
	saveMessageToDB(message)

	server.Mutex.RLock()
	defer server.Mutex.RUnlock()

	for client := range server.Clients {
		if _, ok := client.Groups[message.GroupID]; ok {
			select {
			case client.SendChan <- message:
			default:
				close(client.SendChan)
				delete(server.Clients, client)
			}
		}
	}
}

func (server *ChatServer) sendDirectMessage(message Message) {
	saveMessageToDB(message)

	recipientID, _ := strconv.ParseInt(message.RecipientID, 10, 64)
	senderID, _ := strconv.ParseInt(message.SenderID, 10, 64)

	server.Mutex.RLock()
	defer server.Mutex.RUnlock()

	for client := range server.Clients {
		if client.UserID == recipientID || client.UserID == senderID {
			select {
			case client.SendChan <- message:
			default:
				close(client.SendChan)
				delete(server.Clients, client)
			}
		}
	}
}

func (server *ChatServer) broadcastToAll(message Message) {
	if message.SenderID != "system" {
		saveMessageToDB(message)
	}

	server.Mutex.RLock()
	defer server.Mutex.RUnlock()

	for client := range server.Clients {
		select {
		case client.SendChan <- message:
		default:
			close(client.SendChan)
			delete(server.Clients, client)
		}
	}
}

func saveMessageToDB(message Message) {
	var senderID int64
	if message.SenderID != "system" {
		var err error
		senderID, err = strconv.ParseInt(message.SenderID, 10, 64)
		if err != nil {
			log.Println("Error parsing sender ID:", err)
			return
		}
	}

	var recipientID sql.NullInt64
	if message.RecipientID != "" {
		id, err := strconv.ParseInt(message.RecipientID, 10, 64)
		if err != nil {
			log.Println("Error parsing recipient ID:", err)
		} else {
			recipientID = sql.NullInt64{Int64: id, Valid: true}
		}
	}

	var groupID sql.NullInt64
	if message.GroupID > 0 {
		groupID = sql.NullInt64{Int64: message.GroupID, Valid: true}
	}

	_, err := db.Exec(
		"INSERT INTO messages (message_type, sender_id, recipient_id, group_id, content, url, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		message.MessageType,
		senderID,
		recipientID,
		groupID,
		message.Data.Content,
		message.Data.URL,
		message.Timestamp,
	)
	if err != nil {
		log.Println("Error saving message to database:", err)
	}
}

func (server *ChatServer) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	userID, ok := authenticate(c.Writer, c.Request)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Conn:     conn,
		SendChan: make(chan Message, 256),
		UserID:   userID,
		Groups:   make(map[int64]bool),
	}

	server.Register <- client

	go server.writePump(client)
	server.readPump(client)
}

func (server *ChatServer) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
		server.Unregister <- client
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := client.Conn.WriteJSON(message)
			if err != nil {
				return
			}
		}
	}
}

func (server *ChatServer) readPump(client *Client) {
	defer func() {
		server.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, payload, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var message Message
		if err := json.Unmarshal(payload, &message); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		message.SenderID = strconv.FormatInt(client.UserID, 10)
		message.Timestamp = time.Now()

		server.processMessage(message)
	}
}

func (server *ChatServer) processMessage(message Message) {
	switch message.Type {
	case "bot":
		go server.handleBotMessage(message)
	case "user":
		server.Broadcast <- message
	case "group":
		server.Broadcast <- message
	default:
		log.Println("Unknown message type:", message.Type)
	}
}

func (server *ChatServer) handleBotMessage(message Message) {
	botResponse, err := fetchOllamaResponse(message.Data.Content)
	if err != nil {
		log.Println("Bot response error:", err)
		return
	}

	botMessage := Message{
		Type:        "bot",
		MessageType: "text",
		SenderID:    "system",
		Data: struct {
			URL     string `json:"url"`
			Content string `json:"content"`
		}{
			Content: botResponse,
		},
		Timestamp: time.Now(),
	}

	if message.GroupID > 0 {
		botMessage.GroupID = message.GroupID
	} else if message.RecipientID != "" {
		botMessage.RecipientID = message.SenderID 
	}

	server.Broadcast <- botMessage
}

func fetchOllamaResponse(message string) (string, error) {
	req := LlamaRequest{
		Model: "vietnamese-vision-assistant",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: message,
			},
		},
		Stream: true,
	}
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

func CreateGroupHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	var groupID int64
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO groups (name, description, creator_id) VALUES ($1, $2, $3) RETURNING id",
		req.Name, req.Description, userID,
	).Scan(&groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)",
		groupID, userID, "admin",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator to group"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"group_id":    groupID,
		"name":        req.Name,
		"description": req.Description,
		"creator_id":  userID,
		"message":     "Group created successfully",
	})
}

func InviteToGroupHandler(c *gin.Context) {
	inviterID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req InviteToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var isAdmin bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2 AND role = 'admin')",
		req.GroupID, inviterID,
	).Scan(&isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group admins can invite users"})
		return
	}

	if _, err := users.GetUser(c.Request.Context(), strconv.FormatInt(req.UserID, 10)); err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User to invite does not exist"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User service unavailable"})
		}
		return
	}

	var isMember bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		req.GroupID, req.UserID,
	).Scan(&isMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if isMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this group"})
		return
	}

	_, err = db.Exec(
		"INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)",
		req.GroupID, req.UserID, req.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": req.GroupID,
		"user_id":  req.UserID,
		"role":     req.Role,
		"message":  "User added to group successfully",
	})
}

func GetGroupsHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	rows, err := db.Query(`
		SELECT g.id, g.name, g.description, g.creator_id, g.created_at, gm.role
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	groups := []map[string]interface{}{}
	for rows.Next() {
		var group Group
		var role string
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID, &group.CreatedAt, &role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		groups = append(groups, map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"creator_id":  group.CreatorID,
			"created_at":  group.CreatedAt,
			"role":        role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
}

func GetGroupMembersHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var isMember bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, userID,
	).Scan(&isMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this group"})
		return
	}

	rows, err := db.Query(`
		SELECT user_id, role, joined_at
		FROM group_members
		WHERE group_id = $1
	`, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type member struct {
		id       int64
		role     string
		joinedAt time.Time
	}
	var found []member
	var ids []int64
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.id, &m.role, &m.joinedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		found = append(found, m)
		ids = append(ids, m.id)
	}

	profiles := lookupProfiles(c.Request.Context(), ids)
	members := []map[string]interface{}{}
	for _, m := range found {
		members = append(members, map[string]interface{}{
			"id":        m.id,
			"username":  profiles[m.id].Username,
			"email":     profiles[m.id].Email,
			"role":      m.role,
			"joined_at": m.joinedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": groupID,
		"members":  members,
	})
}

func GetChatHistoryHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var limit int = 50
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}

	var offset int = 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}
	}

	groupIDStr := c.Query("group_id")
	recipientIDStr := c.Query("recipient_id")

	var rows *sql.Rows

	query := `
		SELECT m.id, m.message_type, m.sender_id, m.recipient_id, m.group_id, 
			   m.content, m.url, m.created_at
		FROM messages m
	`
	var args []interface{}

	if groupIDStr != "" {
		groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		var isMember bool
		err = db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
			groupID, userID,
		).Scan(&isMember)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this group"})
			return
		}

		query += "WHERE m.group_id = $1 ORDER BY m.created_at DESC LIMIT $2 OFFSET $3"
		args = append(args, groupID, limit, offset)
	} else if recipientIDStr != "" {
		recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
			return
		}

		query += `
			WHERE (m.sender_id = $1 AND m.recipient_id = $2) 
			   OR (m.sender_id = $2 AND m.recipient_id = $1)
			ORDER BY m.created_at DESC LIMIT $3 OFFSET $4
		`
		args = append(args, userID, recipientID, limit, offset)
	} else {
		query += `
			LEFT JOIN group_members gm ON m.group_id = gm.group_id
			WHERE m.sender_id = $1 
			   OR m.recipient_id = $1 
			   OR (m.group_id IS NOT NULL AND gm.user_id = $1)
			ORDER BY m.created_at DESC LIMIT $2 OFFSET $3
		`
		args = append(args, userID, limit, offset)
	}

	rows, err = db.Query(query, args...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	defer rows.Close()

	messages := []map[string]interface{}{}
	var senderIDs []int64
	for rows.Next() {
		var id int64
		var messageType string
		var senderID sql.NullInt64
		var recipientID sql.NullInt64
		var groupID sql.NullInt64
		var content string
		var url sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &messageType, &senderID, &recipientID, &groupID, &content, &url, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error: " + err.Error()})
			return
		}

		message := map[string]interface{}{
			"id":           id,
			"message_type": messageType,
			"content":      content,
			"created_at":   createdAt,
		}

		if senderID.Valid {
			message["sender_id"] = senderID.Int64
			senderIDs = append(senderIDs, senderID.Int64)
		}
		if recipientID.Valid {
			message["recipient_id"] = recipientID.Int64
		}
		if groupID.Valid {
			message["group_id"] = groupID.Int64
		}
		if url.Valid {
			message["url"] = url.String
		}

		messages = append(messages, message)
	}

	names := lookupProfiles(c.Request.Context(), senderIDs)
	for _, m := range messages {
		if id, ok := m["sender_id"].(int64); ok {
			if p, ok := names[id]; ok {
				m["sender_name"] = p.Username
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
		"offset":   offset,
		"limit":    limit,
	})
}

func LeaveGroupHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if isCreator, err := checkIfGroupCreator(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else if isCreator {
		if err := handleCreatorLeavingGroup(groupID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
		return
	}

	if err := removeUserFromGroup(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left group successfully"})
}

func checkIfGroupCreator(groupID, userID int64) (bool, error) {
	var isCreator bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1 AND creator_id = $2)",
		groupID, userID,
	).Scan(&isCreator)
	return isCreator, err
}

func handleCreatorLeavingGroup(groupID, userID int64) error {
	adminCount, err := countAdminsExcludingUser(groupID, userID)
	if err != nil {
		return fmt.Errorf("Database error")
	}

	if adminCount == 0 {
		memberCount, err := countMembersExcludingUser(groupID, userID)
		if err != nil {
			return fmt.Errorf("Database error")
		}

		if memberCount > 0 {
			return fmt.Errorf("You must promote another member to admin before leaving")
		}

		return deleteGroup(groupID)
	}

	return nil
}

func countAdminsExcludingUser(groupID, userID int64) (int, error) {
	var adminCount int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND role = 'admin' AND user_id != $2",
		groupID, userID,
	).Scan(&adminCount)
	return adminCount, err
}

func countMembersExcludingUser(groupID, userID int64) (int, error) {
	var memberCount int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id != $2",
		groupID, userID,
	).Scan(&memberCount)
	return memberCount, err
}

func deleteGroup(groupID int64) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Transaction error")
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM group_members WHERE group_id = $1", groupID); err != nil {
		return fmt.Errorf("Error removing members")
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM groups WHERE id = $1", groupID); err != nil {
		return fmt.Errorf("Error deleting group")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Transaction commit error")
	}

	return nil
}

func removeUserFromGroup(groupID, userID int64) error {
	_, err := db.Exec("DELETE FROM group_members WHERE group_id = $1 AND user_id = $2", groupID, userID)
	return err
}

func GetUserContactsHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	rows, err := db.Query(`
		SELECT DISTINCT CASE WHEN sender_id = $1 THEN recipient_id ELSE sender_id END
		FROM messages
		WHERE (sender_id = $1 AND recipient_id IS NOT NULL) OR recipient_id = $1
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		if id != userID {
			ids = append(ids, id)
		}
	}

	profiles := lookupProfiles(c.Request.Context(), ids)
	contacts := []map[string]interface{}{}
	for _, id := range ids {
		p, ok := profiles[id]
		if !ok {
			continue // account no longer exists in user-service
		}
		contacts = append(contacts, map[string]interface{}{
			"id":       id,
			"username": p.Username,
			"email":    p.Email,
		})
	}
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i]["username"].(string) < contacts[j]["username"].(string)
	})

	c.JSON(http.StatusOK, gin.H{
		"contacts": contacts,
	})
}

func PromoteGroupMemberHandler(c *gin.Context) {
	adminID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	groupID, err := parseParamID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := parseParamID(c.Param("userId"), "user ID to promote")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isGroupAdmin(groupID, adminID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group admins can promote members"})
		return
	}

	if !isGroupMember(groupID, userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a member of this group"})
		return
	}

	if err := promoteUserToAdmin(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to promote user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": groupID,
		"user_id":  userID,
		"message":  "User promoted to admin successfully",
	})
}

func parseUserID(userIDStr string) (int64, error) {
	return strconv.ParseInt(userIDStr, 10, 64)
}

func parseParamID(param, paramName string) (int64, error) {
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s", paramName)
	}
	return id, nil
}

func isGroupAdmin(groupID, adminID int64) bool {
	var isAdmin bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2 AND role = 'admin')",
		groupID, adminID,
	).Scan(&isAdmin)
	return err == nil && isAdmin
}

func isGroupMember(groupID, userID int64) bool {
	var isMember bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, userID,
	).Scan(&isMember)
	return err == nil && isMember
}

func promoteUserToAdmin(groupID, userID int64) error {
	_, err := db.Exec(
		"UPDATE group_members SET role = 'admin' WHERE group_id = $1 AND user_id = $2",
		groupID, userID,
	)
	return err
}

func RemoveGroupMemberHandler(c *gin.Context) {
	adminID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	groupID, err := parseParamID(c.Param("groupId"), "group ID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := parseParamID(c.Param("userId"), "user ID to remove")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isGroupAdmin(groupID, adminID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only group admins can remove members"})
		return
	}

	if isGroupCreator(groupID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove the group creator"})
		return
	}

	if err := removeUserFromGroup(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": groupID,
		"user_id":  userID,
		"message":  "User removed from group successfully",
	})
}

func isGroupCreator(groupID, userID int64) bool {
	var isCreator bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1 AND creator_id = $2)",
		groupID, userID,
	).Scan(&isCreator)
	return err == nil && isCreator
}

func SearchHandler(c *gin.Context) {
	userID, err := callerID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	groups, err := searchGroups(userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
}

func searchGroups(userID int64, query string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT id, name, description, creator_id
		FROM groups
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY name
		LIMIT 20
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID); err != nil {
			return nil, err
		}

		isMember, err := checkGroupMembership(group.ID, userID)
		if err != nil {
			continue
		}

		groups = append(groups, map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"creator_id":  group.CreatorID,
			"type":        "group",
			"is_member":   isMember,
		})
	}
	return groups, nil
}

func checkGroupMembership(groupID, userID int64) (bool, error) {
	var isMember bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, userID,
	).Scan(&isMember)
	return isMember, err
}

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	chatServer := NewChatServer()
	go chatServer.Run()

	router := setupRouter(chatServer)

	go func() {
		log.Fatal(router.Run(":8080"))
	}()

	startWebSocketServer(chatServer)
}

func setupRouter(chatServer *ChatServer) *gin.Engine {
	router := gin.Default()

	router.Use(corsMiddleware())

	// Sign-up and login happen in user-service; every route below takes its
	// bearer token. Searching all users isn't offered: user data belongs to
	// user-service.
	userRoutes := router.Group("/users", requireAuth())
	{
		userRoutes.GET("/contacts", GetUserContactsHandler)
		userRoutes.GET("/search", SearchHandler)
	}

	groups := router.Group("/groups", requireAuth())
	{
		groups.POST("/create", CreateGroupHandler)
		groups.POST("/invite", InviteToGroupHandler)
		groups.GET("/", GetGroupsHandler)
		groups.GET("/:groupId/members", GetGroupMembersHandler)
		groups.DELETE("/:groupId/leave", LeaveGroupHandler)
		groups.POST("/:groupId/members/:userId/promote", PromoteGroupMemberHandler)
		groups.DELETE("/:groupId/members/:userId", RemoveGroupMemberHandler)
	}

	chat := router.Group("/chat", requireAuth())
	{
		chat.GET("/history", GetChatHistoryHandler)
	}

	router.GET("/ws", func(c *gin.Context) {
		chatServer.HandleWebSocket(c)
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func startWebSocketServer(chatServer *ChatServer) {
	log.Fatal(http.ListenAndServe(":8083", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			handleWebSocketConnection(chatServer, w, r)
			return
		}
		http.NotFound(w, r)
	})))
}

func handleWebSocketConnection(chatServer *ChatServer, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Conn:     conn,
		SendChan: make(chan Message, 256),
		UserID:   userID,
		Groups:   make(map[int64]bool),
	}

	chatServer.Register <- client

	go chatServer.writePump(client)
	chatServer.readPump(client)
}
