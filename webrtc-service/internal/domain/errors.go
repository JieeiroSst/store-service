package domain

import "errors"

var (
	ErrInvalidID      = errors.New("user_id and room_id must be 1-64 printable characters")
	ErrInvalidMessage = errors.New("invalid signaling message")
	ErrPeerNotFound   = errors.New("peer not found in room")
	ErrPeerClosed     = errors.New("peer connection closed")
	ErrPeerBusy       = errors.New("peer send buffer full")
	ErrRoomFull       = errors.New("room is full")
	ErrReplaced       = errors.New("signed in from another session")
)

var errorCodes = []struct {
	err  error
	code string
}{
	{ErrInvalidID, "invalid_id"},
	{ErrInvalidMessage, "invalid_message"},
	{ErrPeerNotFound, "peer_not_found"},
	{ErrPeerClosed, "peer_closed"},
	{ErrPeerBusy, "peer_busy"},
	{ErrRoomFull, "room_full"},
	{ErrReplaced, "replaced"},
}
