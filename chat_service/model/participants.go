package model

import (
	"fmt"
	"time"
)

type TYPE string

const (
	SINGLE TYPE = "single"
	GROUP  TYPE = "group"
)

type Participants struct {
	ID             int          `json:"_id,omitempty" bson:"_id,omitempty"`
	ConversationId int          `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	UsersId        int          `json:"users_id,omitempty" bson:"users_id,omitempty"`
	TYPE           TYPE         `json:"type,omitempty" bson:"type,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt      time.Time    `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Conversation   Conversation `json:"conversation,omitempty" bson:"conversation,omitempty"`
}

func (t TYPE) String() string {
	switch t {
	case SINGLE:
		return "single"
	case GROUP:
		return "group"
	default:
		return fmt.Sprintf("%v", string(t))
	}
}

func ParseType(t string) (c TYPE, err error) {
	types := map[TYPE]struct{}{
		SINGLE: {},
		GROUP:  {},
	}
	cap := TYPE(t)
	_, ok := types[cap]
	if !ok {
		return c, fmt.Errorf(`cannot parse:[%s] as capability`, t)
	}
	return cap, nil
}
