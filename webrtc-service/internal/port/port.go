package port

import "github.com/JIeeiroSst/webrtc-service/internal/domain"

type Peer interface {
	ID() string
	RoomID() string
	Media() domain.MediaState
	SetMedia(domain.MediaState)
	Send(msg domain.Message) error
	Close()
}

type SignalingService interface {
	Join(p Peer) error
	Leave(p Peer)
	Relay(from Peer, msg domain.Message) error
	Room(roomID string) (domain.RoomInfo, error)
}

type ICEService interface {
	Servers(userID string) []domain.ICEServer
}

type RoomRegistry interface {
	Add(p Peer, capacity int) (replaced Peer, others []Peer, err error)
	Remove(p Peer) bool
	Get(roomID, userID string) (Peer, bool)
	Members(roomID string) []Peer
	Stats() (rooms, peers int)
}

type Metrics interface {
	MessageRelayed(t domain.MessageType)
	JoinRejected(reason string)
}
