package model

import "time"

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleMember MemberRole = "member"
)

type Room struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	OwnerName   string    `json:"owner"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Member struct {
	RoomID   uint       `json:"room_id"`
	Username string     `json:"username"`
	Role     MemberRole `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
}
