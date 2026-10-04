package domain

import (
	"strings"
	"time"
)

type HealthStatus string

const (
	HealthPassing  HealthStatus = "passing"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
	HealthAny      HealthStatus = "any"
)

func (s HealthStatus) Valid() bool {
	switch s {
	case HealthPassing, HealthWarning, HealthCritical:
		return true
	}
	return false
}

func (s HealthStatus) severity() int {
	switch s {
	case HealthPassing:
		return 0
	case HealthWarning:
		return 1
	}
	return 2
}

func AggregateStatus(checks []Check) HealthStatus {
	out := HealthPassing
	for _, c := range checks {
		if c.Status.severity() > out.severity() {
			out = c.Status
		}
	}
	return out
}

type RaftIndex struct {
	CreateIndex uint64
	ModifyIndex uint64
}

type Node struct {
	Name    string
	Address string
	Meta    map[string]string
	RaftIndex
}

type Service struct {
	ID      string
	Name    string
	Tags    []string
	Address string
	Port    int
	Meta    map[string]string
	Node    string
	RaftIndex
}

func (s Service) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

type CheckType string

const (
	CheckHTTP   CheckType = "http"
	CheckTCP    CheckType = "tcp"
	CheckTTL    CheckType = "ttl"
	CheckManual CheckType = "manual"
)

const (
	SerfHealthID             = "serfHealth"
	NodeMaintenanceID        = "_node_maintenance"
	ServiceMaintenancePrefix = "_service_maintenance:"
)

type Check struct {
	ID          string
	Name        string
	Node        string
	ServiceID   string
	ServiceName string
	Status      HealthStatus
	Notes       string
	Output      string
	Type        CheckType

	HTTP          string
	Method        string
	Body          string
	Header        map[string][]string
	TLSSkipVerify bool
	TCP           string

	Interval                       time.Duration
	Timeout                        time.Duration
	TTL                            time.Duration
	DeregisterCriticalServiceAfter time.Duration

	CriticalSince time.Time
	TTLExpires    time.Time
	RaftIndex
}

func (c *Check) Validate() error {
	defs := 0
	if c.HTTP != "" {
		defs++
		c.Type = CheckHTTP
	}
	if c.TCP != "" {
		defs++
		c.Type = CheckTCP
	}
	if c.TTL > 0 {
		defs++
		c.Type = CheckTTL
	}
	if defs > 1 {
		return Invalid("check %q: only one of HTTP, TCP or TTL may be set", c.ID)
	}
	if defs == 0 {
		c.Type = CheckManual
	}
	if (c.Type == CheckHTTP || c.Type == CheckTCP) && c.Interval <= 0 {
		return Invalid("check %q: Interval is required for %s checks", c.ID, c.Type)
	}
	if c.Status != "" && !c.Status.Valid() {
		return Invalid("check %q: invalid status %q", c.ID, c.Status)
	}
	return nil
}

type ServiceEntry struct {
	Node    Node
	Service Service
	Checks  []Check
}

type CatalogRegistration struct {
	Node           Node
	Service        *Service
	Checks         []Check
	SkipNodeUpdate bool
}
