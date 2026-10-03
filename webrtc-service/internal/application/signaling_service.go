package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/JIeeiroSst/webrtc-service/internal/port"
)

type SignalingService struct {
	rooms    port.RoomRegistry
	metrics  port.Metrics
	maxPeers int
	now      func() time.Time
}

func NewSignalingService(rooms port.RoomRegistry, metrics port.Metrics, cfg *config.Config) port.SignalingService {
	return &SignalingService{
		rooms:    rooms,
		metrics:  metrics,
		maxPeers: cfg.Room.MaxPeers,
		now:      time.Now,
	}
}

func (s *SignalingService) Join(p port.Peer) error {
	if err := domain.ValidateID(p.ID()); err != nil {
		s.metrics.JoinRejected(domain.ErrorCode(err))
		return err
	}
	if err := domain.ValidateID(p.RoomID()); err != nil {
		s.metrics.JoinRejected(domain.ErrorCode(err))
		return err
	}

	replaced, others, err := s.rooms.Add(p, s.maxPeers)
	if err != nil {
		s.metrics.JoinRejected(domain.ErrorCode(err))
		return err
	}
	if replaced != nil {

		replaced.Send(domain.NewErrorMessage(domain.ErrReplaced))
		replaced.Close()
		s.broadcast(others, domain.Message{Type: domain.MessageUserLeft, RoomID: p.RoomID(), UserID: p.ID()})
	}

	ids := make([]string, 0, len(others))
	for _, o := range others {
		ids = append(ids, o.ID())
	}
	sort.Strings(ids)
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	s.send(p, domain.Message{Type: domain.MessageRoomUsers, RoomID: p.RoomID(), Data: data})

	for _, o := range others {
		s.send(p, mediaStateMessage(o))
	}
	s.broadcast(others, domain.Message{Type: domain.MessageUserJoined, RoomID: p.RoomID(), UserID: p.ID()})

	log.Printf("user %s joined room %s", p.ID(), p.RoomID())
	return nil
}

func (s *SignalingService) Leave(p port.Peer) {
	if !s.rooms.Remove(p) {
		return
	}
	s.broadcast(s.rooms.Members(p.RoomID()), domain.Message{Type: domain.MessageUserLeft, RoomID: p.RoomID(), UserID: p.ID()})
	log.Printf("user %s left room %s", p.ID(), p.RoomID())
}

func (s *SignalingService) Relay(from port.Peer, msg domain.Message) error {

	msg.RoomID = from.RoomID()
	msg.UserID = from.ID()

	var err error
	switch {
	case msg.Type.IsDirected():
		err = s.relayDirected(from, msg)
	case msg.Type == domain.MessageChat:
		err = s.relayChat(from, msg)
	case msg.Type == domain.MessageMediaState:
		err = s.relayMediaState(from, msg)
	default:
		err = fmt.Errorf("%w: type %q cannot be sent by clients", domain.ErrInvalidMessage, msg.Type)
	}
	if err == nil {
		s.metrics.MessageRelayed(msg.Type)
	}
	return err
}

func (s *SignalingService) Room(roomID string) (domain.RoomInfo, error) {
	if err := domain.ValidateID(roomID); err != nil {
		return domain.RoomInfo{}, err
	}
	members := s.rooms.Members(roomID)
	info := domain.RoomInfo{
		RoomID:       roomID,
		Capacity:     s.maxPeers,
		Participants: make([]domain.Participant, 0, len(members)),
	}
	for _, m := range members {
		info.Participants = append(info.Participants, domain.Participant{UserID: m.ID(), Media: m.Media()})
	}
	sort.Slice(info.Participants, func(i, j int) bool {
		return info.Participants[i].UserID < info.Participants[j].UserID
	})
	return info, nil
}

func (s *SignalingService) relayDirected(from port.Peer, msg domain.Message) error {
	if msg.ToUserID == "" || msg.ToUserID == from.ID() {
		return fmt.Errorf("%w: to_user_id must name another peer", domain.ErrInvalidMessage)
	}
	to, err := s.target(msg)
	if err != nil {
		return err
	}
	return s.send(to, msg)
}

func (s *SignalingService) relayChat(from port.Peer, msg domain.Message) error {
	chat, err := domain.ParseChat(msg.Data, s.now())
	if err != nil {
		return err
	}
	if msg.Data, err = json.Marshal(chat); err != nil {
		return err
	}
	if msg.ToUserID == "" {
		s.broadcast(s.others(from), msg)
		return nil
	}
	to, err := s.target(msg)
	if err != nil {
		return err
	}
	return s.send(to, msg)
}

func (s *SignalingService) relayMediaState(from port.Peer, msg domain.Message) error {
	media, err := domain.ParseMediaState(msg.Data)
	if err != nil {
		return err
	}
	from.SetMedia(media)
	s.broadcast(s.others(from), mediaStateMessage(from))
	return nil
}

func (s *SignalingService) target(msg domain.Message) (port.Peer, error) {
	to, ok := s.rooms.Get(msg.RoomID, msg.ToUserID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", domain.ErrPeerNotFound, msg.ToUserID)
	}
	return to, nil
}

func (s *SignalingService) others(p port.Peer) []port.Peer {
	members := s.rooms.Members(p.RoomID())
	out := members[:0]
	for _, m := range members {
		if m.ID() != p.ID() {
			out = append(out, m)
		}
	}
	return out
}

func mediaStateMessage(p port.Peer) domain.Message {
	data, _ := json.Marshal(p.Media())
	return domain.Message{Type: domain.MessageMediaState, RoomID: p.RoomID(), UserID: p.ID(), Data: data}
}

func (s *SignalingService) broadcast(peers []port.Peer, msg domain.Message) {
	for _, p := range peers {
		s.send(p, msg)
	}
}

func (s *SignalingService) send(p port.Peer, msg domain.Message) error {
	err := p.Send(msg)
	if errors.Is(err, domain.ErrPeerBusy) {
		log.Printf("user %s in room %s is not reading, disconnecting", p.ID(), p.RoomID())
		p.Close()
	}
	return err
}
