package model

import "time"

type TranscodeNode struct {
	ID            string
	Addr          string // rtmp://<pod>.<headless-svc>:1935/live
	ActiveStreams int
	MaxStreams    int
	Capacity      int
	LoadPerCore   float64
	LastHeartbeat time.Time
}

func (n TranscodeNode) HasCapacity() bool {
	return n.Capacity > 0
}
