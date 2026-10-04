package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type duration time.Duration

func (d *duration) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			*d = 0
			return nil
		}
		v, err := time.ParseDuration(s)
		if err != nil {
			return domain.Invalid("invalid duration %q", s)
		}
		*d = duration(v)
		return nil
	}
	if string(b) == "null" {
		return nil
	}
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return domain.Invalid("invalid duration %s", b)
	}
	*d = duration(n)
	return nil
}

func durString(d time.Duration) string {
	if d == 0 {
		return ""
	}
	return d.String()
}

type checkDef struct {
	ID        string
	CheckID   string
	Name      string
	ServiceID string
	Status    string
	Notes     string
	Output    string

	HTTP          string
	Method        string
	Body          string
	Header        map[string][]string
	TLSSkipVerify bool
	TCP           string

	Interval                       duration
	Timeout                        duration
	TTL                            duration
	DeregisterCriticalServiceAfter duration

	ScriptArgs        []string
	DockerContainerID string
	GRPC              string
	UDP               string
	H2PING            string
	OSService         string
	AliasNode         string
	AliasService      string
	TLSServerName     string
	TCPUseTLS         bool
}

func (c checkDef) toDomain() (domain.Check, error) {
	switch {
	case len(c.ScriptArgs) > 0 || c.DockerContainerID != "":
		return domain.Check{}, domain.Invalid("script and docker checks are not supported")
	case c.GRPC != "" || c.UDP != "" || c.H2PING != "" || c.OSService != "":
		return domain.Check{}, domain.Invalid("only HTTP, TCP and TTL checks are supported")
	case c.AliasNode != "" || c.AliasService != "":
		return domain.Check{}, domain.Invalid("alias checks are not supported")
	case c.TLSServerName != "" || c.TCPUseTLS:
		return domain.Check{}, domain.Invalid("TLSServerName and TCPUseTLS are not supported")
	}
	id := c.CheckID
	if id == "" {
		id = c.ID
	}
	return domain.Check{
		ID:                             id,
		Name:                           c.Name,
		ServiceID:                      c.ServiceID,
		Status:                         domain.HealthStatus(c.Status),
		Notes:                          c.Notes,
		Output:                         c.Output,
		HTTP:                           c.HTTP,
		Method:                         c.Method,
		Body:                           c.Body,
		Header:                         c.Header,
		TLSSkipVerify:                  c.TLSSkipVerify,
		TCP:                            c.TCP,
		Interval:                       time.Duration(c.Interval),
		Timeout:                        time.Duration(c.Timeout),
		TTL:                            time.Duration(c.TTL),
		DeregisterCriticalServiceAfter: time.Duration(c.DeregisterCriticalServiceAfter),
	}, nil
}

type catalogCheckIn struct {
	CheckID    string
	Name       string
	Status     string
	Notes      string
	Output     string
	ServiceID  string
	Definition checkDef
}

func (c catalogCheckIn) toDomain() (domain.Check, error) {
	def := c.Definition
	def.CheckID, def.Name, def.Status, def.Notes, def.Output, def.ServiceID =
		c.CheckID, c.Name, c.Status, c.Notes, c.Output, c.ServiceID
	return def.toDomain()
}

type checkDefinitionOut struct {
	HTTP                           string              `json:",omitempty"`
	Header                         map[string][]string `json:",omitempty"`
	Method                         string              `json:",omitempty"`
	Body                           string              `json:",omitempty"`
	TLSSkipVerify                  bool                `json:",omitempty"`
	TCP                            string              `json:",omitempty"`
	Interval                       string              `json:",omitempty"`
	Timeout                        string              `json:",omitempty"`
	DeregisterCriticalServiceAfter string              `json:",omitempty"`
}

type healthCheckOut struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
	ServiceTags []string
	Type        string
	Interval    string `json:",omitempty"`
	Timeout     string `json:",omitempty"`
	TTL         string `json:",omitempty"`
	Definition  checkDefinitionOut
	CreateIndex uint64
	ModifyIndex uint64
}

func checkOut(c domain.Check, tags []string) healthCheckOut {
	typ := string(c.Type)
	if c.Type == domain.CheckManual {
		typ = ""
	}
	if tags == nil {
		tags = []string{}
	}
	return healthCheckOut{
		Node:        c.Node,
		CheckID:     c.ID,
		Name:        c.Name,
		Status:      string(c.Status),
		Notes:       c.Notes,
		Output:      c.Output,
		ServiceID:   c.ServiceID,
		ServiceName: c.ServiceName,
		ServiceTags: tags,
		Type:        typ,
		Interval:    durString(c.Interval),
		Timeout:     durString(c.Timeout),
		TTL:         durString(c.TTL),
		Definition: checkDefinitionOut{
			HTTP:                           c.HTTP,
			Header:                         c.Header,
			Method:                         c.Method,
			Body:                           c.Body,
			TLSSkipVerify:                  c.TLSSkipVerify,
			TCP:                            c.TCP,
			Interval:                       durString(c.Interval),
			Timeout:                        durString(c.Timeout),
			DeregisterCriticalServiceAfter: durString(c.DeregisterCriticalServiceAfter),
		},
		CreateIndex: c.CreateIndex,
		ModifyIndex: c.ModifyIndex,
	}
}

func checksOut(cs []domain.Check) []healthCheckOut {
	out := make([]healthCheckOut, 0, len(cs))
	for _, c := range cs {
		out = append(out, checkOut(c, nil))
	}
	return out
}

type serviceRegistration struct {
	ID      string
	Name    string
	Tags    []string
	Port    int
	Address string
	Meta    map[string]string
	Kind    string
	Check   *checkDef
	Checks  []checkDef
}

func (r serviceRegistration) toDomain() (domain.Service, []domain.Check, error) {
	if r.Kind != "" && r.Kind != "typical" {
		return domain.Service{}, nil, domain.Invalid("service kind %q is not supported", r.Kind)
	}
	defs := r.Checks
	if r.Check != nil {
		defs = append([]checkDef{*r.Check}, defs...)
	}
	checks := make([]domain.Check, 0, len(defs))
	for _, d := range defs {
		c, err := d.toDomain()
		if err != nil {
			return domain.Service{}, nil, err
		}
		checks = append(checks, c)
	}
	return domain.Service{ID: r.ID, Name: r.Name, Tags: r.Tags, Port: r.Port, Address: r.Address, Meta: r.Meta}, checks, nil
}

type weights struct{ Passing, Warning int }

type agentServiceOut struct {
	Kind              string
	ID                string
	Service           string
	Tags              []string
	Meta              map[string]string
	Port              int
	Address           string
	Weights           weights
	EnableTagOverride bool
	Datacenter        string
	CreateIndex       uint64
	ModifyIndex       uint64
}

func serviceOut(s domain.Service, dc string) agentServiceOut {
	tags, meta := s.Tags, s.Meta
	if tags == nil {
		tags = []string{}
	}
	if meta == nil {
		meta = map[string]string{}
	}
	return agentServiceOut{
		ID: s.ID, Service: s.Name, Tags: tags, Meta: meta, Port: s.Port, Address: s.Address,
		Weights: weights{Passing: 1, Warning: 1}, Datacenter: dc,
		CreateIndex: s.CreateIndex, ModifyIndex: s.ModifyIndex,
	}
}

type catalogServiceIn struct {
	ID      string
	Service string
	Tags    []string
	Address string
	Meta    map[string]string
	Port    int
}

type nodeOut struct {
	ID              string
	Node            string
	Address         string
	Datacenter      string
	TaggedAddresses map[string]string
	Meta            map[string]string
	CreateIndex     uint64
	ModifyIndex     uint64
}

func toNodeOut(n domain.Node, dc string) nodeOut {
	meta := n.Meta
	if meta == nil {
		meta = map[string]string{}
	}
	return nodeOut{
		Node: n.Name, Address: n.Address, Datacenter: dc,
		TaggedAddresses: map[string]string{"lan": n.Address, "wan": n.Address},
		Meta:            meta, CreateIndex: n.CreateIndex, ModifyIndex: n.ModifyIndex,
	}
}

type catalogServiceOut struct {
	ID              string
	Node            string
	Address         string
	Datacenter      string
	TaggedAddresses map[string]string
	NodeMeta        map[string]string
	ServiceID       string
	ServiceName     string
	ServiceAddress  string
	ServiceTags     []string
	ServiceMeta     map[string]string
	ServicePort     int
	ServiceWeights  weights
	CreateIndex     uint64
	ModifyIndex     uint64
}

func catalogServiceRow(e domain.ServiceEntry, dc string) catalogServiceOut {
	n := toNodeOut(e.Node, dc)
	s := serviceOut(e.Service, dc)
	return catalogServiceOut{
		Node: n.Node, Address: n.Address, Datacenter: dc, TaggedAddresses: n.TaggedAddresses, NodeMeta: n.Meta,
		ServiceID: s.ID, ServiceName: s.Service, ServiceAddress: s.Address, ServiceTags: s.Tags,
		ServiceMeta: s.Meta, ServicePort: s.Port, ServiceWeights: s.Weights,
		CreateIndex: s.CreateIndex, ModifyIndex: s.ModifyIndex,
	}
}

type serviceEntryOut struct {
	Node    nodeOut
	Service agentServiceOut
	Checks  []healthCheckOut
}

func serviceEntry(e domain.ServiceEntry, dc string) serviceEntryOut {
	out := serviceEntryOut{Node: toNodeOut(e.Node, dc), Service: serviceOut(e.Service, dc)}
	out.Checks = make([]healthCheckOut, 0, len(e.Checks))
	for _, c := range e.Checks {
		var tags []string
		if c.ServiceID != "" {
			tags = e.Service.Tags
		}
		out.Checks = append(out.Checks, checkOut(c, tags))
	}
	return out
}

type kvOut struct {
	LockIndex   uint64
	Key         string
	Flags       uint64
	Value       []byte
	Session     string `json:",omitempty"`
	CreateIndex uint64
	ModifyIndex uint64
}

func toKVOut(p domain.KVPair) kvOut {
	return kvOut{
		LockIndex: p.LockIndex, Key: p.Key, Flags: p.Flags, Value: p.Value, Session: p.Session,
		CreateIndex: p.CreateIndex, ModifyIndex: p.ModifyIndex,
	}
}

type lockDelay time.Duration

func (d *lockDelay) UnmarshalJSON(b []byte) error {
	var v duration
	if err := v.UnmarshalJSON(b); err != nil {
		return err
	}
	if len(b) > 0 && b[0] != '"' && time.Duration(v) < 1000 {
		v = duration(time.Duration(v) * time.Second)
	}
	*d = lockDelay(v)
	return nil
}

type sessionIn struct {
	Name          string
	Node          string
	LockDelay     *lockDelay
	Behavior      string
	TTL           duration
	Checks        *[]string
	NodeChecks    *[]string
	ServiceChecks []json.RawMessage
}

func (s sessionIn) toDomain() (domain.Session, error) {
	if len(s.ServiceChecks) > 0 {
		return domain.Session{}, domain.Invalid("ServiceChecks are not supported")
	}
	out := domain.Session{
		Name:      s.Name,
		Node:      s.Node,
		Behavior:  domain.SessionBehavior(s.Behavior),
		TTL:       time.Duration(s.TTL),
		LockDelay: domain.DefaultLockDelay,
	}
	if s.LockDelay != nil {
		out.LockDelay = time.Duration(*s.LockDelay)
	}
	switch {
	case s.Checks == nil && s.NodeChecks == nil:
		out.NodeChecks = []string{domain.SerfHealthID}
	default:
		if s.Checks != nil {
			out.NodeChecks = append(out.NodeChecks, *s.Checks...)
		}
		if s.NodeChecks != nil {
			out.NodeChecks = append(out.NodeChecks, *s.NodeChecks...)
		}
	}
	return out, nil
}

type sessionOut struct {
	ID            string
	Name          string
	Node          string
	LockDelay     time.Duration
	Behavior      string
	TTL           string
	Checks        []string
	NodeChecks    []string
	ServiceChecks []struct{}
	CreateIndex   uint64
	ModifyIndex   uint64
}

func toSessionOut(s domain.Session) sessionOut {
	checks := s.NodeChecks
	if checks == nil {
		checks = []string{}
	}
	return sessionOut{
		ID: s.ID, Name: s.Name, Node: s.Node, LockDelay: s.LockDelay, Behavior: string(s.Behavior),
		TTL: durString(s.TTL), Checks: checks, NodeChecks: checks, ServiceChecks: []struct{}{},
		CreateIndex: s.CreateIndex, ModifyIndex: s.ModifyIndex,
	}
}

type intentionIn struct {
	SourceName      string
	DestinationName string
	Action          string
	Description     string
	Meta            map[string]string
	Permissions     []json.RawMessage
}

func (i intentionIn) toDomain() (domain.Intention, error) {
	if len(i.Permissions) > 0 {
		return domain.Intention{}, domain.Invalid("L7 Permissions are not supported, use Action")
	}
	return domain.Intention{
		SourceName: i.SourceName, DestinationName: i.DestinationName,
		Action: domain.IntentionAction(i.Action), Description: i.Description, Meta: i.Meta,
	}, nil
}

type intentionOut struct {
	ID              string
	Description     string `json:",omitempty"`
	SourceNS        string
	SourceName      string
	DestinationNS   string
	DestinationName string
	SourceType      string
	Action          string
	Meta            map[string]string `json:",omitempty"`
	Precedence      int
	CreateIndex     uint64
	ModifyIndex     uint64
}

func toIntentionOut(i domain.Intention) intentionOut {
	return intentionOut{
		ID: i.ID, Description: i.Description, SourceNS: "default", SourceName: i.SourceName,
		DestinationNS: "default", DestinationName: i.DestinationName, SourceType: "consul",
		Action: string(i.Action), Meta: i.Meta, Precedence: i.Precedence,
		CreateIndex: i.CreateIndex, ModifyIndex: i.ModifyIndex,
	}
}

func intentionsOut(in []domain.Intention) []intentionOut {
	out := make([]intentionOut, 0, len(in))
	for _, i := range in {
		out = append(out, toIntentionOut(i))
	}
	return out
}

func leaderAddr(addr string) string { return fmt.Sprintf("%s:8300", addr) }
