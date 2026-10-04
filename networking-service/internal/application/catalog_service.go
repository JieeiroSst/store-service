package application

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type AgentInfo struct {
	NodeName   string
	Address    string
	Datacenter string
	Version    string
}

type ServiceSummary struct {
	Name           string
	Tags           []string
	Instances      int
	Nodes          int
	ChecksPassing  int
	ChecksWarning  int
	ChecksCritical int
}

type CatalogService struct {
	store   port.CatalogStore
	blocker *Blocker
	self    AgentInfo
	now     func() time.Time
}

func NewCatalogService(store port.CatalogStore, blocker *Blocker, cfg *config.Config) *CatalogService {
	return &CatalogService{
		store:   store,
		blocker: blocker,
		self: AgentInfo{
			NodeName:   cfg.Agent.NodeName,
			Address:    cfg.Agent.NodeAddress,
			Datacenter: cfg.Agent.Datacenter,
			Version:    config.Version,
		},
		now: time.Now,
	}
}

func (c *CatalogService) Self() AgentInfo { return c.self }

func (c *CatalogService) RegisterSelf() error {
	return c.store.EnsureRegistration(domain.CatalogRegistration{
		Node: domain.Node{Name: c.self.NodeName, Address: c.self.Address},
		Checks: []domain.Check{{
			ID:     domain.SerfHealthID,
			Name:   "Serf Health Status",
			Status: domain.HealthPassing,
			Output: "Agent alive and reachable",
			Type:   domain.CheckManual,
		}},
	})
}

type AgentServiceRegistration struct {
	Service               domain.Service
	Checks                []domain.Check
	ReplaceExistingChecks bool
}

func (c *CatalogService) AgentRegisterService(reg AgentServiceRegistration) error {
	svc := reg.Service
	if svc.Name == "" {
		return domain.Invalid("missing service name")
	}
	if svc.ID == "" {
		svc.ID = svc.Name
	}
	if svc.Port < 0 || svc.Port > 65535 {
		return domain.Invalid("invalid port %d", svc.Port)
	}
	checks := make([]domain.Check, 0, len(reg.Checks))
	for i, chk := range reg.Checks {
		chk.ServiceID = svc.ID
		if chk.ID == "" {
			chk.ID = "service:" + svc.ID
			if len(reg.Checks) > 1 {
				chk.ID = fmt.Sprintf("service:%s:%d", svc.ID, i+1)
			}
		}
		if chk.Name == "" {
			chk.Name = fmt.Sprintf("Service '%s' check", svc.Name)
		}
		if err := c.prepareCheck(&chk); err != nil {
			return err
		}
		checks = append(checks, chk)
	}

	if err := c.store.EnsureRegistration(domain.CatalogRegistration{
		Node:           domain.Node{Name: c.self.NodeName, Address: c.self.Address},
		Service:        &svc,
		Checks:         checks,
		SkipNodeUpdate: true,
	}); err != nil {
		return err
	}
	if !reg.ReplaceExistingChecks {
		return nil
	}
	keep := map[string]bool{}
	for _, chk := range checks {
		keep[chk.ID] = true
	}
	_, all, err := c.store.Checks()
	if err != nil {
		return err
	}
	for _, chk := range all {
		if chk.Node == c.self.NodeName && chk.ServiceID == svc.ID && !keep[chk.ID] &&
			chk.ID != domain.ServiceMaintenancePrefix+svc.ID {
			if err := c.store.DeregisterCheck(chk.Node, chk.ID); err != nil && !errors.Is(err, domain.ErrNotFound) {
				return err
			}
		}
	}
	return nil
}

func (c *CatalogService) AgentDeregisterService(id string) error {
	if err := c.store.DeregisterService(c.self.NodeName, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("unknown service ID %q: %w", id, err)
		}
		return err
	}
	return nil
}

func (c *CatalogService) AgentServices() ([]domain.Service, error) {
	_, _, services, err := c.store.Node(c.self.NodeName)
	return services, err
}

func (c *CatalogService) AgentService(id string) (*domain.Service, error) {
	services, err := c.AgentServices()
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("unknown service ID %q: %w", id, domain.ErrNotFound)
}

func (c *CatalogService) AgentChecks() ([]domain.Check, error) {
	_, all, err := c.store.Checks()
	if err != nil {
		return nil, err
	}
	var out []domain.Check
	for _, chk := range all {
		if chk.Node == c.self.NodeName {
			out = append(out, chk)
		}
	}
	return out, nil
}

func (c *CatalogService) AgentRegisterCheck(chk domain.Check) error {
	if chk.Name == "" {
		return domain.Invalid("missing check name")
	}
	if chk.ID == "" {
		chk.ID = chk.Name
	}
	if err := c.prepareCheck(&chk); err != nil {
		return err
	}
	return c.store.EnsureRegistration(domain.CatalogRegistration{
		Node:           domain.Node{Name: c.self.NodeName, Address: c.self.Address},
		Checks:         []domain.Check{chk},
		SkipNodeUpdate: true,
	})
}

func (c *CatalogService) AgentDeregisterCheck(id string) error {
	if err := c.store.DeregisterCheck(c.self.NodeName, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("unknown check ID %q: %w", id, err)
		}
		return err
	}
	return nil
}

func (c *CatalogService) UpdateTTL(checkID string, status domain.HealthStatus, output string) error {
	if !status.Valid() {
		return domain.Invalid("invalid status %q", status)
	}
	checks, err := c.AgentChecks()
	if err != nil {
		return err
	}
	for _, chk := range checks {
		if chk.ID != checkID {
			continue
		}
		if chk.Type != domain.CheckTTL {
			return domain.Invalid("CheckID %q does not have associated TTL", checkID)
		}
		_, err := c.store.UpdateCheck(chk.Node, chk.ID, status, output, c.now().Add(chk.TTL))
		return err
	}
	return fmt.Errorf("unknown check ID %q: %w", checkID, domain.ErrNotFound)
}

func (c *CatalogService) ServiceMaintenance(serviceID string, enable bool, reason string) error {
	svc, err := c.AgentService(serviceID)
	if err != nil {
		return err
	}
	id := domain.ServiceMaintenancePrefix + svc.ID
	if !enable {
		if err := c.store.DeregisterCheck(c.self.NodeName, id); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return nil
	}
	return c.store.EnsureRegistration(domain.CatalogRegistration{
		Node:           domain.Node{Name: c.self.NodeName},
		SkipNodeUpdate: true,
		Checks: []domain.Check{{
			ID: id, Name: "Service Maintenance Mode", ServiceID: svc.ID,
			Status: domain.HealthCritical, Notes: maintenanceNotes(reason), Type: domain.CheckManual,
		}},
	})
}

func (c *CatalogService) NodeMaintenance(enable bool, reason string) error {
	if !enable {
		if err := c.store.DeregisterCheck(c.self.NodeName, domain.NodeMaintenanceID); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return nil
	}
	return c.store.EnsureRegistration(domain.CatalogRegistration{
		Node:           domain.Node{Name: c.self.NodeName},
		SkipNodeUpdate: true,
		Checks: []domain.Check{{
			ID: domain.NodeMaintenanceID, Name: "Node Maintenance Mode",
			Status: domain.HealthCritical, Notes: maintenanceNotes(reason), Type: domain.CheckManual,
		}},
	})
}

func maintenanceNotes(reason string) string {
	if reason == "" {
		return "Maintenance mode is enabled for this service, but no reason was provided. This is a default message."
	}
	return reason
}

func (c *CatalogService) prepareCheck(chk *domain.Check) error {
	if err := chk.Validate(); err != nil {
		return err
	}
	if chk.Type == domain.CheckTTL {
		chk.TTLExpires = c.now().Add(chk.TTL)
	}
	return nil
}

func (c *CatalogService) Register(reg domain.CatalogRegistration) error {
	if reg.Node.Name == "" {
		return domain.Invalid("missing node name")
	}
	if reg.Node.Address == "" && !reg.SkipNodeUpdate {
		return domain.Invalid("missing node address")
	}
	if reg.Service != nil {
		if reg.Service.Name == "" {
			return domain.Invalid("missing service name")
		}
		if reg.Service.ID == "" {
			reg.Service.ID = reg.Service.Name
		}
	}
	for i := range reg.Checks {
		chk := &reg.Checks[i]
		if chk.ID == "" {
			chk.ID = chk.Name
		}
		if chk.ID == "" {
			return domain.Invalid("missing check ID")
		}
		if chk.Name == "" {
			chk.Name = chk.ID
		}
		if err := c.prepareCheck(chk); err != nil {
			return err
		}
	}
	return c.store.EnsureRegistration(reg)
}

func (c *CatalogService) Deregister(node, serviceID, checkID string) error {
	if node == "" {
		return domain.Invalid("missing node name")
	}
	var err error
	switch {
	case serviceID != "":
		err = c.store.DeregisterService(node, serviceID)
	case checkID != "":
		err = c.store.DeregisterCheck(node, checkID)
	default:
		err = c.store.DeregisterNode(node)
	}
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	return err
}

func (c *CatalogService) Nodes(ctx context.Context, q QueryOptions) ([]domain.Node, QueryMeta, error) {
	var out []domain.Node
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		idx, nodes, err := c.store.Nodes()
		out = out[:0]
		for _, n := range nodes {
			if matchNodeMeta(n.Meta, q.NodeMeta) {
				out = append(out, n)
			}
		}
		return idx, err
	})
	return out, m, err
}

func (c *CatalogService) Node(ctx context.Context, q QueryOptions, name string) (*domain.Node, []domain.Service, QueryMeta, error) {
	var (
		node     *domain.Node
		services []domain.Service
	)
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		var (
			idx uint64
			err error
		)
		idx, node, services, err = c.store.Node(name)
		return idx, err
	})
	return node, services, m, err
}

func (c *CatalogService) Services(ctx context.Context, q QueryOptions) (map[string][]string, QueryMeta, error) {
	var out map[string][]string
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		idx, services, err := c.store.Services()
		if err != nil {
			return 0, err
		}
		var nodeMeta map[string]map[string]string
		if len(q.NodeMeta) > 0 {
			_, nodes, err := c.store.Nodes()
			if err != nil {
				return 0, err
			}
			nodeMeta = map[string]map[string]string{}
			for _, n := range nodes {
				nodeMeta[n.Name] = n.Meta
			}
		}
		tags := map[string]map[string]struct{}{}
		for _, s := range services {
			if nodeMeta != nil && !matchNodeMeta(nodeMeta[s.Node], q.NodeMeta) {
				continue
			}
			if tags[s.Name] == nil {
				tags[s.Name] = map[string]struct{}{}
			}
			for _, t := range s.Tags {
				tags[s.Name][t] = struct{}{}
			}
		}
		out = make(map[string][]string, len(tags))
		for name, set := range tags {
			out[name] = slices.Sorted(maps.Keys(set))
		}
		return idx, nil
	})
	return out, m, err
}

func (c *CatalogService) ServiceNodes(ctx context.Context, q QueryOptions, name string, tags []string, passingOnly bool) ([]domain.ServiceEntry, QueryMeta, error) {
	var out []domain.ServiceEntry
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		idx, entries, err := c.store.ServiceNodes(name)
		out = filterEntries(entries, tags, q.NodeMeta, passingOnly)
		return idx, err
	})
	return out, m, err
}

func filterEntries(entries []domain.ServiceEntry, tags []string, nodeMeta map[string]string, passingOnly bool) []domain.ServiceEntry {
	out := make([]domain.ServiceEntry, 0, len(entries))
	for _, e := range entries {
		if !matchNodeMeta(e.Node.Meta, nodeMeta) {
			continue
		}
		if !slices.ContainsFunc(tags, func(t string) bool { return !e.Service.HasTag(t) }) &&
			(!passingOnly || domain.AggregateStatus(e.Checks) == domain.HealthPassing) {
			out = append(out, e)
		}
	}
	return out
}

func (c *CatalogService) ServiceChecks(ctx context.Context, q QueryOptions, service string) ([]domain.Check, QueryMeta, error) {
	return c.checksWhere(ctx, q, func(chk domain.Check) bool { return chk.ServiceName == service })
}

func (c *CatalogService) NodeChecks(ctx context.Context, q QueryOptions, node string) ([]domain.Check, QueryMeta, error) {
	return c.checksWhere(ctx, q, func(chk domain.Check) bool { return chk.Node == node })
}

func (c *CatalogService) ChecksInState(ctx context.Context, q QueryOptions, state domain.HealthStatus) ([]domain.Check, QueryMeta, error) {
	if state != domain.HealthAny && !state.Valid() {
		return nil, QueryMeta{}, domain.Invalid("invalid state %q", state)
	}
	return c.checksWhere(ctx, q, func(chk domain.Check) bool {
		return state == domain.HealthAny || chk.Status == state
	})
}

func (c *CatalogService) checksWhere(ctx context.Context, q QueryOptions, keep func(domain.Check) bool) ([]domain.Check, QueryMeta, error) {
	var out []domain.Check
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		idx, all, err := c.store.Checks()
		out = make([]domain.Check, 0, len(all))
		for _, chk := range all {
			if keep(chk) {
				out = append(out, chk)
			}
		}
		return idx, err
	})
	return out, m, err
}

func (c *CatalogService) ServiceSummaries(ctx context.Context, q QueryOptions) ([]ServiceSummary, QueryMeta, error) {
	var out []ServiceSummary
	m, err := c.blocker.Query(ctx, c.store, q, func() (uint64, error) {
		idx, services, err := c.store.Services()
		if err != nil {
			return 0, err
		}
		_, checks, err := c.store.Checks()
		if err != nil {
			return 0, err
		}
		out = summarize(services, checks)
		return idx, nil
	})
	return out, m, err
}

func summarize(services []domain.Service, checks []domain.Check) []ServiceSummary {
	byName := map[string]*ServiceSummary{}
	tags := map[string]map[string]struct{}{}
	nodes := map[string]map[string]struct{}{}
	svcName := map[string]string{}
	for _, s := range services {
		sum := byName[s.Name]
		if sum == nil {
			sum = &ServiceSummary{Name: s.Name}
			byName[s.Name] = sum
			tags[s.Name] = map[string]struct{}{}
			nodes[s.Name] = map[string]struct{}{}
		}
		sum.Instances++
		nodes[s.Name][s.Node] = struct{}{}
		for _, t := range s.Tags {
			tags[s.Name][t] = struct{}{}
		}
		svcName[s.Node+"/"+s.ID] = s.Name
	}
	for _, chk := range checks {
		sum := byName[svcName[chk.Node+"/"+chk.ServiceID]]
		if chk.ServiceID == "" || sum == nil {
			continue
		}
		switch chk.Status {
		case domain.HealthPassing:
			sum.ChecksPassing++
		case domain.HealthWarning:
			sum.ChecksWarning++
		default:
			sum.ChecksCritical++
		}
	}
	out := make([]ServiceSummary, 0, len(byName))
	for _, name := range slices.Sorted(maps.Keys(byName)) {
		sum := byName[name]
		sum.Tags = slices.Sorted(maps.Keys(tags[name]))
		sum.Nodes = len(nodes[name])
		out = append(out, *sum)
	}
	return out
}
