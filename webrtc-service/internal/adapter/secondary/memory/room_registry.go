package memory

import (
	"sync"

	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/JIeeiroSst/webrtc-service/internal/port"
)

type RoomRegistry struct {
	mu    sync.RWMutex
	rooms map[string]map[string]port.Peer
}

func NewRoomRegistry() port.RoomRegistry {
	return &RoomRegistry{rooms: make(map[string]map[string]port.Peer)}
}

func (r *RoomRegistry) Add(p port.Peer, capacity int) (port.Peer, []port.Peer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	room := r.rooms[p.RoomID()]
	replaced := room[p.ID()]
	if replaced == nil && capacity > 0 && len(room) >= capacity {
		return nil, nil, domain.ErrRoomFull
	}
	if room == nil {
		room = make(map[string]port.Peer)
		r.rooms[p.RoomID()] = room
	}

	others := make([]port.Peer, 0, len(room))
	for id, member := range room {
		if id != p.ID() {
			others = append(others, member)
		}
	}
	room[p.ID()] = p
	return replaced, others, nil
}

func (r *RoomRegistry) Remove(p port.Peer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	room := r.rooms[p.RoomID()]
	if room == nil || room[p.ID()] != p {
		return false
	}
	delete(room, p.ID())
	if len(room) == 0 {
		delete(r.rooms, p.RoomID())
	}
	return true
}

func (r *RoomRegistry) Get(roomID, userID string) (port.Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.rooms[roomID][userID]
	return p, ok
}

func (r *RoomRegistry) Members(roomID string) []port.Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room := r.rooms[roomID]
	members := make([]port.Peer, 0, len(room))
	for _, p := range room {
		members = append(members, p)
	}
	return members
}

func (r *RoomRegistry) Stats() (rooms, peers int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, room := range r.rooms {
		peers += len(room)
	}
	return len(r.rooms), peers
}
