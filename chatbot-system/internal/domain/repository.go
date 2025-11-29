package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetManagedUsers(ctx context.Context, managerID int64) ([]*User, error)
	GetAdvisedUser(ctx context.Context, advisorID int64) (*User, error)
	Update(ctx context.Context, user *User) error
}

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) error
	GetByID(ctx context.Context, id int64) (*Conversation, error)
	GetByUsers(ctx context.Context, user1ID, user2ID int64) (*Conversation, error)
	GetUserConversations(ctx context.Context, userID int64) ([]*Conversation, error)
	Update(ctx context.Context, conv *Conversation) error
}

type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id int64) (*Message, error)
	GetConversationMessages(ctx context.Context, conversationID int64, limit, offset int) ([]*Message, error)
	GetConversationHistory(ctx context.Context, conversationID int64) ([]*Message, error)
}
