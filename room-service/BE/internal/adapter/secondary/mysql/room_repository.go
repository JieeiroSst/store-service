package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository { return &RoomRepository{db: db} }

const roomSelect = `rooms.id, rooms.name, rooms.owner AS owner_name, rooms.created_at,
	(SELECT COUNT(*) FROM room_members WHERE room_members.room_id = rooms.id) AS member_count`

type roomView struct {
	ID          uint
	Name        string
	OwnerName   string
	MemberCount int
	CreatedAt   time.Time
}

func (v roomView) toModel() model.Room {
	return model.Room{ID: v.ID, Name: v.Name, OwnerName: v.OwnerName, MemberCount: v.MemberCount, CreatedAt: v.CreatedAt}
}

func (r *RoomRepository) Create(ctx context.Context, name, owner string) (model.Room, error) {
	var id uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		room := roomRow{Name: name, Owner: owner}
		if err := tx.Create(&room).Error; err != nil {
			return err
		}
		id = room.ID
		return tx.Create(&memberRow{RoomID: room.ID, Username: owner, Role: string(model.RoleOwner)}).Error
	})
	if err != nil {
		return model.Room{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *RoomRepository) FindByID(ctx context.Context, id uint) (model.Room, error) {
	var v roomView
	res := r.db.WithContext(ctx).Table("rooms").Select(roomSelect).Where("rooms.id = ?", id).Limit(1).Scan(&v)
	if res.Error != nil {
		return model.Room{}, res.Error
	}
	if res.RowsAffected == 0 {
		return model.Room{}, port.ErrNotFound
	}
	return v.toModel(), nil
}

func (r *RoomRepository) ListByMember(ctx context.Context, username string) ([]model.Room, error) {
	var views []roomView
	err := r.db.WithContext(ctx).Table("rooms").Select(roomSelect).
		Joins("JOIN room_members m ON m.room_id = rooms.id AND m.username = ?", username).
		Order("rooms.id DESC").Scan(&views).Error
	if err != nil {
		return nil, err
	}
	rooms := make([]model.Room, len(views))
	for i, v := range views {
		rooms[i] = v.toModel()
	}
	return rooms, nil
}

func (r *RoomRepository) Members(ctx context.Context, roomID uint) ([]model.Member, error) {
	var rows []memberRow
	if err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Order("joined_at, username").Find(&rows).Error; err != nil {
		return nil, err
	}
	members := make([]model.Member, len(rows))
	for i, row := range rows {
		members[i] = row.toModel()
	}
	return members, nil
}

func (r *RoomRepository) Member(ctx context.Context, roomID uint, username string) (model.Member, error) {
	var row memberRow
	err := r.db.WithContext(ctx).Where("room_id = ? AND username = ?", roomID, username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Member{}, port.ErrNotFound
	}
	if err != nil {
		return model.Member{}, err
	}
	return row.toModel(), nil
}

func (r *RoomRepository) AddMember(ctx context.Context, m model.Member) error {
	err := r.db.WithContext(ctx).Create(&memberRow{RoomID: m.RoomID, Username: m.Username, Role: string(m.Role)}).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return port.ErrAlreadyMember
	}
	return err
}

func (r *RoomRepository) RemoveMember(ctx context.Context, roomID uint, username string) error {
	return r.db.WithContext(ctx).Where("room_id = ? AND username = ?", roomID, username).Delete(&memberRow{}).Error
}
