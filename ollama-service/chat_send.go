package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxMessageContent = 4000

var (
	errMessageTarget   = errors.New("exactly one of recipient_id or group_id is required")
	errMessageEmpty    = errors.New("content or url is required")
	errMessageTooLong  = errors.New("content is too long")
	errMessageSelf     = errors.New("cannot send a message to yourself")
	errNotGroupMember  = errors.New("not a member of this group")
	errMessageKind     = errors.New("message_type must be text, image or file")
	groupMembershipFor = isGroupMember
	groupMembersOf     = loadGroupMemberIDs
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = 50 * time.Second
	wsMaxMessage = 64 << 10
)

func mustID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

type SendMessageRequest struct {
	RecipientID int64  `json:"recipient_id"`
	GroupID     int64  `json:"group_id"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	URL         string `json:"url"`
}

func buildOutgoingMessage(senderID int64, req SendMessageRequest, now time.Time) (Message, error) {
	var msg Message
	if (req.RecipientID > 0) == (req.GroupID > 0) {
		return msg, errMessageTarget
	}
	content := strings.TrimSpace(req.Content)
	url := strings.TrimSpace(req.URL)
	if content == "" && url == "" {
		return msg, errMessageEmpty
	}
	if len(content) > maxMessageContent {
		return msg, errMessageTooLong
	}
	kind := req.MessageType
	if kind == "" {
		kind = "text"
	}
	if kind != "text" && kind != "image" && kind != "file" {
		return msg, errMessageKind
	}
	if req.RecipientID == senderID {
		return msg, errMessageSelf
	}
	if req.GroupID > 0 && !groupMembershipFor(req.GroupID, senderID) {
		return msg, errNotGroupMember
	}

	msg.MessageType = kind
	msg.SenderID = strconv.FormatInt(senderID, 10)
	msg.Data.Content = content
	msg.Data.URL = url
	msg.Timestamp = now
	if req.GroupID > 0 {
		msg.Type = "group"
		msg.GroupID = req.GroupID
	} else {
		msg.Type = "user"
		msg.RecipientID = strconv.FormatInt(req.RecipientID, 10)
	}
	return msg, nil
}

func SendMessageHandler(server *ChatServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := callerID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msg, err := buildOutgoingMessage(userID, req, time.Now())
		switch {
		case errors.Is(err, errNotGroupMember):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case err != nil:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		server.Broadcast <- msg

		c.JSON(http.StatusAccepted, gin.H{
			"message_type": msg.MessageType,
			"sender_id":    userID,
			"recipient_id": req.RecipientID,
			"group_id":     req.GroupID,
			"content":      msg.Data.Content,
			"url":          msg.Data.URL,
			"created_at":   msg.Timestamp,
		})
	}
}
