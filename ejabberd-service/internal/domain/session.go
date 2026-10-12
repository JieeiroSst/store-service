package domain

import (
	"net"
	"time"
)

type SID struct {
	Time time.Time
	Seq  uint64
}

func (s SID) Less(o SID) bool {
	if s.Time.Equal(o.Time) {
		return s.Seq < o.Seq
	}
	return s.Time.Before(o.Time)
}

type SessionInfo struct {
	IP         net.IP
	Conn       string
	AuthModule string
}

type Session struct {
	SID      SID
	USR      LJID
	Priority *int
	Info     SessionInfo
}

func (s Session) Available() bool { return s.Priority != nil }

type ResourceConflict string

const (
	ConflictAcceptNew   ResourceConflict = "acceptnew"
	ConflictCloseOld    ResourceConflict = "closeold"
	ConflictCloseNew    ResourceConflict = "closenew"
	ConflictSetResource ResourceConflict = "setresource"
)
