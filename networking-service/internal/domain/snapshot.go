package domain

import "time"

type Snapshot struct {
	Index      uint64
	TakenAt    time.Time
	Nodes      []Node
	Services   []Service
	Checks     []Check
	KV         []KVPair
	Sessions   []Session
	Intentions []Intention
}

type Stats struct {
	Nodes      int
	Services   int
	Instances  int
	Checks     map[HealthStatus]int
	Keys       int
	Sessions   int
	Intentions int
}
