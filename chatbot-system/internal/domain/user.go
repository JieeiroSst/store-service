package domain

import "time"

type UserRole string

const (
	RoleManager UserRole = "manager"
	RoleAdvisor UserRole = "advisor"
	RoleUser    UserRole = "user"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	ManagerID *int64    `json:"manager_id,omitempty"` // For users managed by a manager
	AdvisorID *int64    `json:"advisor_id,omitempty"` // For users assigned to an advisor
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) CanChatWith(targetUser *User) bool {
	switch u.Role {
	case RoleManager:
		// Manager can chat with their managed users
		return targetUser.ManagerID != nil && *targetUser.ManagerID == u.ID
	case RoleAdvisor:
		// Advisor can chat with their assigned user
		return targetUser.AdvisorID != nil && *targetUser.AdvisorID == u.ID
	case RoleUser:
		// User can chat with their manager or advisor
		if u.ManagerID != nil && *u.ManagerID == targetUser.ID {
			return true
		}
		if u.AdvisorID != nil && *u.AdvisorID == targetUser.ID {
			return true
		}
	}
	return false
}
