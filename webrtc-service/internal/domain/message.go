package domain

import (
	"encoding/json"
	"errors"
	"unicode/utf8"
)

type MessageType string

const (
	MessageOffer        MessageType = "offer"
	MessageAnswer       MessageType = "answer"
	MessageICECandidate MessageType = "ice_candidate"

	MessageChat       MessageType = "chat"
	MessageMediaState MessageType = "media_state"

	MessageUserJoined MessageType = "user_joined"
	MessageUserLeft   MessageType = "user_left"
	MessageRoomUsers  MessageType = "room_users"
	MessageError      MessageType = "error"
)

const MaxIDLength = 64

type Message struct {
	Type     MessageType     `json:"type"`
	RoomID   string          `json:"room_id,omitempty"`
	UserID   string          `json:"user_id,omitempty"`
	ToUserID string          `json:"to_user_id,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

func (t MessageType) IsDirected() bool {
	switch t {
	case MessageOffer, MessageAnswer, MessageICECandidate:
		return true
	}
	return false
}

func ValidateID(id string) error {
	if id == "" || len(id) > MaxIDLength || !utf8.ValidString(id) {
		return ErrInvalidID
	}
	for _, r := range id {
		if r < 0x20 || r == 0x7f {
			return ErrInvalidID
		}
	}
	return nil
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorMessage(err error) Message {
	data, _ := json.Marshal(ErrorPayload{Code: ErrorCode(err), Message: err.Error()})
	return Message{Type: MessageError, Data: data}
}

func ErrorCode(err error) string {
	for _, c := range errorCodes {
		if errors.Is(err, c.err) {
			return c.code
		}
	}
	return "internal"
}
