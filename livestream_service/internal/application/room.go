package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/google/uuid"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

type roomUsecase struct {
	rooms port.RoomRepository
}

func NewRoomUsecase(rooms port.RoomRepository) port.RoomUsecase {
	return &roomUsecase{rooms: rooms}
}

func (u *roomUsecase) CreateRoom(ctx context.Context, input model.CreateRoomInput) (*model.Room, error) {
	id := uuid.NewString()
	key, err := generateStreamKey()
	if err != nil {
		return nil, fmt.Errorf("generate stream key: %w", err)
	}

	room := &model.Room{
		ID:          id,
		OwnerUserID: input.OwnerUserID,
		Slug:        slugify(input.Title) + "-" + id[:8],
		Title:       input.Title,
		Description: input.Description,
		StreamKey:   key,
		Status:      model.RoomStatusOffline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.rooms.Create(ctx, room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return room, nil
}

func (u *roomUsecase) GetRoom(ctx context.Context, roomID string) (*model.Room, error) {
	return u.rooms.GetByID(ctx, roomID)
}

func (u *roomUsecase) ListRooms(ctx context.Context, live bool) ([]model.Room, error) {
	return u.rooms.List(ctx, live)
}

func (u *roomUsecase) RegenerateStreamKey(ctx context.Context, roomID string) (string, error) {
	key, err := generateStreamKey()
	if err != nil {
		return "", fmt.Errorf("generate stream key: %w", err)
	}
	if err := u.rooms.UpdateStreamKey(ctx, roomID, key); err != nil {
		return "", fmt.Errorf("update stream key: %w", err)
	}
	return key, nil
}

func generateStreamKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(buf), nil
}

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonSlugChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "room"
	}
	return s
}
