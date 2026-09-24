package application

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
)

const (
	maxRoomNameLen = 100
	maxUsernameLen = 64
)

type RoomService struct {
	rooms  port.RoomRepository
	events port.EventPublisher
}

func NewRoomService(rooms port.RoomRepository, events port.EventPublisher) *RoomService {
	return &RoomService{rooms: rooms, events: events}
}

func (s *RoomService) CreateRoom(ctx context.Context, actor model.User, name string) (model.Room, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxRoomNameLen {
		return model.Room{}, port.ErrInvalidInput
	}
	return s.rooms.Create(ctx, name, actor.Username)
}

func (s *RoomService) ListRooms(ctx context.Context, actor model.User) ([]model.Room, error) {
	return s.rooms.ListByMember(ctx, actor.Username)
}

func (s *RoomService) GetRoom(ctx context.Context, actor model.User, roomID uint) (model.Room, error) {
	if _, err := s.requireMember(ctx, actor, roomID); err != nil {
		return model.Room{}, err
	}
	return s.rooms.FindByID(ctx, roomID)
}

func (s *RoomService) ListMembers(ctx context.Context, actor model.User, roomID uint) ([]model.Member, error) {
	if _, err := s.requireMember(ctx, actor, roomID); err != nil {
		return nil, err
	}
	return s.rooms.Members(ctx, roomID)
}

func (s *RoomService) AddMember(ctx context.Context, actor model.User, roomID uint, username string) (model.Member, error) {
	username = strings.TrimSpace(username)
	if username == "" || utf8.RuneCountInString(username) > maxUsernameLen || strings.ContainsAny(username, " \t\r\n") {
		return model.Member{}, port.ErrInvalidInput
	}
	me, err := s.requireMember(ctx, actor, roomID)
	if err != nil {
		return model.Member{}, err
	}
	if me.Role != model.RoleOwner {
		return model.Member{}, port.ErrForbidden
	}

	m := model.Member{RoomID: roomID, Username: username, Role: model.RoleMember}
	if err := s.rooms.AddMember(ctx, m); err != nil {
		return model.Member{}, err
	}
	added, err := s.rooms.Member(ctx, roomID, username)
	if err != nil {
		return model.Member{}, err
	}
	s.publish(ctx, model.Event{Type: model.EventMemberAdded, RoomID: roomID, Member: &added})
	return added, nil
}

func (s *RoomService) RemoveMember(ctx context.Context, actor model.User, roomID uint, username string) error {
	me, err := s.requireMember(ctx, actor, roomID)
	if err != nil {
		return err
	}
	if username != actor.Username && me.Role != model.RoleOwner {
		return port.ErrForbidden
	}
	target, err := s.rooms.Member(ctx, roomID, username)
	if err != nil {
		return err
	}
	if target.Role == model.RoleOwner {
		return port.ErrForbidden
	}
	if err := s.rooms.RemoveMember(ctx, roomID, username); err != nil {
		return err
	}
	s.publish(ctx, model.Event{Type: model.EventMemberRemoved, RoomID: roomID, Member: &target})
	return nil
}

func (s *RoomService) requireMember(ctx context.Context, actor model.User, roomID uint) (model.Member, error) {
	return requireMember(ctx, s.rooms, actor, roomID)
}

func (s *RoomService) publish(ctx context.Context, e model.Event) {
	if err := s.events.Publish(ctx, e); err != nil {
		log.Printf("publish %s: %v", e.Type, err)
	}
}

func requireMember(ctx context.Context, rooms port.RoomRepository, actor model.User, roomID uint) (model.Member, error) {
	m, err := rooms.Member(ctx, roomID, actor.Username)
	if err == nil {
		return m, nil
	}
	if !errors.Is(err, port.ErrNotFound) {
		return model.Member{}, err
	}
	if _, err := rooms.FindByID(ctx, roomID); err != nil {
		return model.Member{}, err
	}
	return model.Member{}, port.ErrForbidden
}
