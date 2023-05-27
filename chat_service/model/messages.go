package model

import (
	"fmt"
	"time"
)

type MESSAGETYPE string

const (
	TEXT  MESSAGETYPE = "text"
	IMAGE MESSAGETYPE = "image"
	VIDEO MESSAGETYPE = "video"
	AUDIO MESSAGETYPE = "audio"
)

type Messages struct {
	ID          int         `json:"_id,omitempty" bson:"_id,omitempty"`
	Guid        string      `json:"guid,omitempty" bson:"guid,omitempty"`
	SenderId    int         `json:"sender_id,omitempty" bson:"sender_id,omitempty"`
	MessageType MESSAGETYPE `json:"message_type,omitempty" bson:"message_type,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt   time.Time   `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

func (m MESSAGETYPE) String() string {
	switch m {
	case TEXT:
		return "text"
	case IMAGE:
		return "image"
	case VIDEO:
		return "video"
	case AUDIO:
		return "audio"
	default:
		return fmt.Sprintf("%v", string(m))
	}
}

func ParseMessageType(s string) (c MESSAGETYPE, err error) {
	status := map[MESSAGETYPE]struct{}{
		TEXT:  {},
		IMAGE: {},
		VIDEO: {},
		AUDIO: {},
	}
	cap := MESSAGETYPE(s)
	_, ok := status[cap]
	if !ok {
		return c, fmt.Errorf(`cannot parse:[%s] as capability`, s)
	}
	return cap, nil
}
