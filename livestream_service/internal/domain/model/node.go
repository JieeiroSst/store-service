package model

import "time"

type TranscodeNode struct {
	ID            string
	Addr          string // rtmp://<pod>.<headless-svc>:1935/live
	HTTPAddr      string // http://<pod>.<headless-svc>:8080 - for edge->node admin calls (force-unpublish)
	ActiveStreams int
	MaxStreams    int
	Capacity      int
	LoadPerCore   float64
	LastHeartbeat time.Time
}

func (n TranscodeNode) HasCapacity() bool {
	return n.Capacity > 0
}
