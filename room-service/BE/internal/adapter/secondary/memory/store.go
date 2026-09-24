package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
)

type Store struct {
	mu       sync.Mutex
	rooms    map[uint]model.Room
	members  map[uint][]model.Member
	messages []model.Message
	nextRoom uint
	nextMsg  uint
}

func NewStore() *Store {
	return &Store{rooms: map[uint]model.Room{}, members: map[uint][]model.Member{}}
}

func (s *Store) Create(_ context.Context, name, owner string) (model.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextRoom++
	r := model.Room{ID: s.nextRoom, Name: name, OwnerName: owner, CreatedAt: time.Now()}
	s.rooms[r.ID] = r
	s.members[r.ID] = []model.Member{{RoomID: r.ID, Username: owner, Role: model.RoleOwner, JoinedAt: time.Now()}}
	return s.view(r), nil
}

func (s *Store) view(r model.Room) model.Room {
	r.MemberCount = len(s.members[r.ID])
	return r
}

func (s *Store) FindByID(_ context.Context, id uint) (model.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rooms[id]
	if !ok {
		return model.Room{}, port.ErrNotFound
	}
	return s.view(r), nil
}

func (s *Store) ListByMember(_ context.Context, username string) ([]model.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []model.Room
	for id, ms := range s.members {
		for _, m := range ms {
			if m.Username == username {
				out = append(out, s.view(s.rooms[id]))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (s *Store) Members(_ context.Context, roomID uint) ([]model.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.Member(nil), s.members[roomID]...), nil
}

func (s *Store) Member(_ context.Context, roomID uint, username string) (model.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.members[roomID] {
		if m.Username == username {
			return m, nil
		}
	}
	return model.Member{}, port.ErrNotFound
}

func (s *Store) AddMember(_ context.Context, m model.Member) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.members[m.RoomID] {
		if x.Username == m.Username {
			return port.ErrAlreadyMember
		}
	}
	m.JoinedAt = time.Now()
	s.members[m.RoomID] = append(s.members[m.RoomID], m)
	return nil
}

func (s *Store) RemoveMember(_ context.Context, roomID uint, username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ms := s.members[roomID]
	for i, m := range ms {
		if m.Username == username {
			s.members[roomID] = append(ms[:i:i], ms[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *Store) CreateMessage(_ context.Context, m *model.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextMsg++
	m.ID = s.nextMsg
	s.messages = append(s.messages, *m)
	return nil
}

func (s *Store) ListMessages(_ context.Context, roomID, beforeID uint, limit int) ([]model.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []model.Message
	for i := len(s.messages) - 1; i >= 0 && len(out) < limit; i-- {
		m := s.messages[i]
		if m.RoomID == roomID && (beforeID == 0 || m.ID < beforeID) {
			out = append([]model.Message{m}, out...)
		}
	}
	return out, nil
}

type Messages struct{ *Store }

func (m Messages) Create(ctx context.Context, msg *model.Message) error {
	return m.Store.CreateMessage(ctx, msg)
}

func (m Messages) List(ctx context.Context, roomID, beforeID uint, limit int) ([]model.Message, error) {
	return m.Store.ListMessages(ctx, roomID, beforeID, limit)
}
