package mysql

import (
	"time"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
)

type roomRow struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Owner     string `gorm:"type:varchar(64) COLLATE utf8mb4_bin;not null"`
	CreatedAt time.Time
}

func (roomRow) TableName() string { return "rooms" }

type memberRow struct {
	RoomID   uint      `gorm:"primaryKey;autoIncrement:false"`
	Username string    `gorm:"primaryKey;index;type:varchar(64) COLLATE utf8mb4_bin"`
	Role     string    `gorm:"size:16;not null"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

func (memberRow) TableName() string { return "room_members" }

type messageRow struct {
	ID        uint   `gorm:"primaryKey"`
	RoomID    uint   `gorm:"not null;index"`
	Username  string `gorm:"type:varchar(64) COLLATE utf8mb4_bin;not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
}

func (messageRow) TableName() string { return "messages" }

func (m messageRow) toModel() model.Message {
	return model.Message{ID: m.ID, RoomID: m.RoomID, Username: m.Username, Content: m.Content, CreatedAt: m.CreatedAt}
}

func (m memberRow) toModel() model.Member {
	return model.Member{RoomID: m.RoomID, Username: m.Username, Role: model.MemberRole(m.Role), JoinedAt: m.JoinedAt}
}
