package dns

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/miekg/dns"
)

type nopMetrics struct{}

func (nopMetrics) CheckRun(domain.CheckType, domain.HealthStatus) {}
func (nopMetrics) SessionInvalidated(string)                      {}
func (nopMetrics) BlockingQueryWoken()                            {}

func startDNS(t *testing.T) (string, *application.CatalogService) {
	t.Helper()
	cfg := config.FromEnv()
	cfg.Agent.NodeName = "agent-1"
	cfg.Agent.NodeAddress = "10.1.0.5"
	cfg.Agent.Datacenter = "dc1"
	cfg.DNS.Domain = "consul"
	cfg.DNS.Port = "0"
	store := memory.NewStore()
	catalog := application.NewCatalogService(store, application.NewBlocker(cfg, nopMetrics{}), cfg)
	if err := catalog.RegisterSelf(); err != nil {
		t.Fatal(err)
	}
	srv := NewServer(NewHandler(catalog, cfg), cfg)
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Stop(context.Background()) })
	return srv.Addr(), catalog
}

func query(t *testing.T, addr, name string, qtype uint16) *dns.Msg {
	t.Helper()
	m := new(dns.Msg)
	m.SetQuestion(name, qtype)
	c := &dns.Client{Timeout: 2 * time.Second}
	var (
		r   *dns.Msg
		err error
	)
	for range 20 {
		if r, _, err = c.Exchange(m, addr); err == nil {
			return r
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("query %s: %v", name, err)
	return nil
}

func register(t *testing.T, c *application.CatalogService, id, addr string, tags []string, status domain.HealthStatus) {
	t.Helper()
	err := c.Register(domain.CatalogRegistration{
		Node:    domain.Node{Name: "node-" + id, Address: addr},
		Service: &domain.Service{ID: id, Name: "web", Port: 8080, Tags: tags},
		Checks:  []domain.Check{{ID: "c-" + id, ServiceID: id, Status: status}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceLookup(t *testing.T) {
	addr, catalog := startDNS(t)
	register(t, catalog, "web-1", "10.0.0.1", []string{"primary"}, domain.HealthPassing)
	register(t, catalog, "web-2", "10.0.0.2", nil, domain.HealthWarning)
	register(t, catalog, "web-3", "10.0.0.3", nil, domain.HealthCritical)

	r := query(t, addr, "web.service.consul.", dns.TypeA)
	if r.Rcode != dns.RcodeSuccess || len(r.Answer) != 2 {
		t.Fatalf("A answers = %v (critical instance must be left out)", r.Answer)
	}
	if !r.Authoritative {
		t.Fatal("answer not authoritative")
	}

	r = query(t, addr, "primary.web.service.dc1.consul.", dns.TypeA)
	if len(r.Answer) != 1 || r.Answer[0].(*dns.A).A.String() != "10.0.0.1" {
		t.Fatalf("tagged answers = %v", r.Answer)
	}

	r = query(t, addr, "_web._primary.service.consul.", dns.TypeSRV)
	if len(r.Answer) != 1 {
		t.Fatalf("SRV answers = %v", r.Answer)
	}
	srv := r.Answer[0].(*dns.SRV)
	if srv.Port != 8080 || srv.Target != "node-web-1.node.dc1.consul." {
		t.Fatalf("SRV = %v", srv)
	}
	if len(r.Extra) != 1 || r.Extra[0].(*dns.A).A.String() != "10.0.0.1" {
		t.Fatalf("SRV extra = %v", r.Extra)
	}
}

func TestNodeAndMisses(t *testing.T) {
	addr, _ := startDNS(t)
	r := query(t, addr, "agent-1.node.consul.", dns.TypeA)
	if len(r.Answer) != 1 || r.Answer[0].(*dns.A).A.String() != "10.1.0.5" {
		t.Fatalf("node answer = %v", r.Answer)
	}
	if r := query(t, addr, "nope.service.consul.", dns.TypeA); r.Rcode != dns.RcodeNameError {
		t.Fatalf("rcode = %s, want NXDOMAIN", dns.RcodeToString[r.Rcode])
	}
	if r := query(t, addr, "web.service.dc9.consul.", dns.TypeA); r.Rcode != dns.RcodeNameError {
		t.Fatalf("other dc rcode = %s", dns.RcodeToString[r.Rcode])
	}
	if r := query(t, addr, "example.com.", dns.TypeA); r.Rcode != dns.RcodeRefused {
		t.Fatalf("foreign domain rcode = %s, want REFUSED", dns.RcodeToString[r.Rcode])
	}
}

func TestSRVTargetForServiceAddress(t *testing.T) {
	addr, catalog := startDNS(t)
	err := catalog.Register(domain.CatalogRegistration{
		Node:    domain.Node{Name: "k8s-node", Address: "10.0.0.9"},
		Service: &domain.Service{ID: "api-pod", Name: "api", Address: "10.244.1.7", Port: 9000},
		Checks:  []domain.Check{{ID: "c", ServiceID: "api-pod", Status: domain.HealthPassing}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := query(t, addr, "api.service.consul.", dns.TypeSRV)
	if len(r.Answer) != 1 {
		t.Fatalf("SRV = %v", r.Answer)
	}
	target := r.Answer[0].(*dns.SRV).Target
	if target != "0af40107.addr.dc1.consul." {
		t.Fatalf("target = %s", target)
	}
	r = query(t, addr, target, dns.TypeA)
	if len(r.Answer) != 1 || r.Answer[0].(*dns.A).A.String() != "10.244.1.7" {
		t.Fatalf("addr lookup = %v", r.Answer)
	}
}
