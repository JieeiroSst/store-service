package http

import (
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type memberOut struct {
	Name   string
	Addr   string
	Port   int
	Tags   map[string]string
	Status int
}

func (h *Handler) member() memberOut {
	return memberOut{
		Name: h.self.NodeName, Addr: h.self.Address, Port: 8301, Status: 1,
		Tags: map[string]string{"role": "consul", "dc": h.self.Datacenter, "build": h.self.Version},
	}
}

func (h *Handler) AgentSelf(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	h.ok(w, r, nil, map[string]any{
		"Config": map[string]any{
			"Datacenter": h.self.Datacenter,
			"NodeName":   h.self.NodeName,
			"Server":     true,
			"Version":    h.self.Version,
		},
		"Member": h.member(),
		"Meta":   map[string]string{},
	})
}

func (h *Handler) AgentMembers(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	h.ok(w, r, nil, []memberOut{h.member()})
}

func (h *Handler) AgentServices(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	services, err := h.catalog.AgentServices()
	if err != nil {
		h.fail(w, err)
		return
	}
	out := map[string]agentServiceOut{}
	for _, s := range services {
		out[s.ID] = serviceOut(s, h.self.Datacenter)
	}
	h.ok(w, r, nil, out)
}

func (h *Handler) AgentService(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	s, err := h.catalog.AgentService(r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, serviceOut(*s, h.self.Datacenter))
}

func (h *Handler) AgentRegisterService(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in serviceRegistration
	if err := decode(r, &in); err != nil {
		h.fail(w, err)
		return
	}
	svc, checks, err := in.toDomain()
	if err == nil {
		err = h.catalog.AgentRegisterService(application.AgentServiceRegistration{
			Service: svc, Checks: checks, ReplaceExistingChecks: flag(r, "replace-existing-checks"),
		})
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentDeregisterService(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	if err := h.catalog.AgentDeregisterService(r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentServiceMaintenance(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	enable, err := enableParam(r)
	if err == nil {
		err = h.catalog.ServiceMaintenance(r.PathValue("id"), enable, r.URL.Query().Get("reason"))
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentNodeMaintenance(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	enable, err := enableParam(r)
	if err == nil {
		err = h.catalog.NodeMaintenance(enable, r.URL.Query().Get("reason"))
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func enableParam(r *http.Request) (bool, error) {
	switch r.URL.Query().Get("enable") {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, domain.Invalid("Missing value for enable")
}

func (h *Handler) AgentChecks(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	checks, err := h.catalog.AgentChecks()
	if err != nil {
		h.fail(w, err)
		return
	}
	out := map[string]healthCheckOut{}
	for _, c := range checks {
		out[c.ID] = checkOut(c, nil)
	}
	h.ok(w, r, nil, out)
}

func (h *Handler) AgentRegisterCheck(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in checkDef
	if err := decode(r, &in); err != nil {
		h.fail(w, err)
		return
	}
	c, err := in.toDomain()
	if err == nil {
		err = h.catalog.AgentRegisterCheck(c)
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentDeregisterCheck(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	if err := h.catalog.AgentDeregisterCheck(r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentCheckTTL(status domain.HealthStatus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.write(w, r) {
			return
		}
		if err := h.catalog.UpdateTTL(r.PathValue("id"), status, r.URL.Query().Get("note")); err != nil {
			h.fail(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) AgentCheckUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in struct{ Status, Output string }
	if err := decode(r, &in); err != nil {
		h.fail(w, err)
		return
	}
	if err := h.catalog.UpdateTTL(r.PathValue("id"), domain.HealthStatus(in.Status), in.Output); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AgentHealthByName(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	name := r.PathValue("name")
	checks, err := h.catalog.AgentChecks()
	if err != nil {
		h.fail(w, err)
		return
	}
	services, err := h.catalog.AgentServices()
	if err != nil {
		h.fail(w, err)
		return
	}
	type row struct {
		AggregatedStatus string
		Service          agentServiceOut
		Checks           []healthCheckOut
	}
	out := []row{}
	worst := domain.HealthPassing
	for _, s := range services {
		if s.Name != name {
			continue
		}
		var mine []domain.Check
		for _, c := range checks {
			if c.ServiceID == s.ID || c.ServiceID == "" {
				mine = append(mine, c)
			}
		}
		st := domain.AggregateStatus(mine)
		worst = domain.AggregateStatus([]domain.Check{{Status: worst}, {Status: st}})
		out = append(out, row{AggregatedStatus: string(st), Service: serviceOut(s, h.self.Datacenter), Checks: checksOut(mine)})
	}
	code := http.StatusOK
	switch {
	case len(out) == 0:
		code = http.StatusNotFound
	case worst == domain.HealthWarning:
		code = http.StatusTooManyRequests
	case worst == domain.HealthCritical:
		code = http.StatusServiceUnavailable
	}
	h.reply(w, r, nil, code, out)
}
