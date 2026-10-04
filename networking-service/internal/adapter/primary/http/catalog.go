package http

import (
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func (h *Handler) CatalogRegister(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in struct {
		Node           string
		Address        string
		NodeMeta       map[string]string
		Service        *catalogServiceIn
		Check          *catalogCheckIn
		Checks         []catalogCheckIn
		SkipNodeUpdate bool
	}
	if err := decode(r, &in); err != nil {
		h.fail(w, err)
		return
	}
	reg := domain.CatalogRegistration{
		Node:           domain.Node{Name: in.Node, Address: in.Address, Meta: in.NodeMeta},
		SkipNodeUpdate: in.SkipNodeUpdate,
	}
	if s := in.Service; s != nil {
		reg.Service = &domain.Service{ID: s.ID, Name: s.Service, Tags: s.Tags, Address: s.Address, Meta: s.Meta, Port: s.Port}
	}
	checks := in.Checks
	if in.Check != nil {
		checks = append([]catalogCheckIn{*in.Check}, checks...)
	}
	for _, c := range checks {
		dc, err := c.toDomain()
		if err != nil {
			h.fail(w, err)
			return
		}
		reg.Checks = append(reg.Checks, dc)
	}
	if err := h.catalog.Register(reg); err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) CatalogDeregister(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in struct{ Node, ServiceID, CheckID string }
	if err := decode(r, &in); err != nil {
		h.fail(w, err)
		return
	}
	if err := h.catalog.Deregister(in.Node, in.ServiceID, in.CheckID); err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) CatalogDatacenters(w http.ResponseWriter, r *http.Request) {
	h.ok(w, r, nil, []string{h.self.Datacenter})
}

func (h *Handler) CatalogNodes(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	nodes, m, err := h.catalog.Nodes(r.Context(), q)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]nodeOut, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, toNodeOut(n, h.self.Datacenter))
	}
	h.ok(w, r, &m, out)
}

func (h *Handler) CatalogNode(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	node, services, m, err := h.catalog.Node(r.Context(), q, r.PathValue("node"))
	if err != nil {
		h.fail(w, err)
		return
	}
	if node == nil {
		h.ok(w, r, &m, nil)
		return
	}
	out := struct {
		Node     nodeOut
		Services map[string]agentServiceOut
	}{Node: toNodeOut(*node, h.self.Datacenter), Services: map[string]agentServiceOut{}}
	for _, s := range services {
		out.Services[s.ID] = serviceOut(s, h.self.Datacenter)
	}
	h.ok(w, r, &m, out)
}

func (h *Handler) CatalogServices(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	services, m, err := h.catalog.Services(r.Context(), q)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, services)
}

func (h *Handler) CatalogService(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	entries, m, err := h.catalog.ServiceNodes(r.Context(), q, r.PathValue("service"), r.URL.Query()["tag"], false)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]catalogServiceOut, 0, len(entries))
	for _, e := range entries {
		out = append(out, catalogServiceRow(e, h.self.Datacenter))
	}
	h.ok(w, r, &m, out)
}

func (h *Handler) HealthService(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	entries, m, err := h.catalog.ServiceNodes(r.Context(), q, r.PathValue("service"), r.URL.Query()["tag"], flag(r, "passing"))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]serviceEntryOut, 0, len(entries))
	for _, e := range entries {
		out = append(out, serviceEntry(e, h.self.Datacenter))
	}
	h.ok(w, r, &m, out)
}

func (h *Handler) HealthServiceChecks(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	checks, m, err := h.catalog.ServiceChecks(r.Context(), q, r.PathValue("service"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, checksOut(checks))
}

func (h *Handler) HealthNode(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	checks, m, err := h.catalog.NodeChecks(r.Context(), q, r.PathValue("node"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, checksOut(checks))
}

func (h *Handler) HealthState(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	checks, m, err := h.catalog.ChecksInState(r.Context(), q, domain.HealthStatus(r.PathValue("state")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, checksOut(checks))
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) (q queryOpts, ok bool) {
	if !h.read(w, r) {
		return q, false
	}
	q, err := queryOptions(r)
	if err != nil {
		h.fail(w, err)
		return q, false
	}
	return q, true
}
