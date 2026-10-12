package config

import (
	"net"
	"os"
	"testing"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

func TestParseDefaultConfig(t *testing.T) {
	raw, err := os.ReadFile("../../conf/ejabberd.yml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Hosts) != 1 || cfg.Hosts[0] != "localhost" {
		t.Fatalf("hosts = %v", cfg.Hosts)
	}
	var c2s *Listener
	for i, l := range cfg.Listen {
		if l.Port == 5222 {
			c2s = &cfg.Listen[i]
		}
	}
	if c2s == nil || c2s.Module != "ejabberd_c2s" || !c2s.StartTLSRequired || c2s.MaxStanzaSize != 262144 || c2s.Shaper != "c2s_shaper" {
		t.Fatalf("c2s listener = %+v", c2s)
	}
	if !cfg.Modules.Roster.Enabled || !cfg.Modules.Roster.Versioning {
		t.Fatalf("roster = %+v", cfg.Modules.Roster)
	}
	if !cfg.Modules.Offline.Enabled || cfg.Modules.Offline.AccessMaxUserMessages != "max_user_offline_messages" {
		t.Fatalf("offline = %+v", cfg.Modules.Offline)
	}
	if cfg.Modules.Register.IPAccess != "trusted_network" {
		t.Fatalf("register = %+v", cfg.Modules.Register)
	}
	if cfg.Shapers["normal"] != (domain.ShaperSpec{Rate: 3000, Burst: 20000}) || cfg.Shapers["fast"].Rate != 100000 {
		t.Fatalf("shapers = %+v", cfg.Shapers)
	}

	m := cfg.Matcher()
	local := domain.MatchJID(domain.MustJID("bob", "localhost", ""))
	remote := domain.MatchJID(domain.MustJID("bob", "example.org", ""))
	if !m.MatchRule("local", local) || m.MatchRule("local", remote) {
		t.Fatal("local rule mismatch")
	}
	if !m.MatchRule("c2s", local) {
		t.Fatal("c2s must allow regular users")
	}
	if !m.MatchRule("trusted_network", domain.MatchIP(net.ParseIP("127.0.0.1"))) ||
		m.MatchRule("trusted_network", domain.MatchIP(net.ParseIP("10.1.2.3"))) {
		t.Fatal("trusted_network mismatch")
	}
	if v := m.MatchShaper("max_user_offline_messages", local); v != "100" {
		t.Fatalf("max_user_offline_messages = %s", v)
	}
	if v := m.MatchShaper("max_user_sessions", local); v != "10" {
		t.Fatalf("max_user_sessions = %s", v)
	}
	if v := m.MatchShaper("c2s_shaper", local); v != "normal" {
		t.Fatalf("c2s_shaper = %s", v)
	}
	if len(cfg.APIPermissions) != 3 {
		t.Fatalf("api_permissions = %d", len(cfg.APIPermissions))
	}
}

func TestAccessRuleListForm(t *testing.T) {
	cfg, err := Parse([]byte(`
acl:
  admin:
    user: admin@localhost
access_rules:
  configure:
    - deny: blocked
    - allow: admin
  everyone:
    - allow
`))
	if err != nil {
		t.Fatal(err)
	}
	m := cfg.Matcher()
	if !m.MatchRule("configure", domain.MatchJID(domain.MustJID("admin", "localhost", ""))) {
		t.Fatal("admin must be allowed")
	}
	if m.MatchRule("configure", domain.MatchJID(domain.MustJID("bob", "localhost", ""))) {
		t.Fatal("bob must be denied")
	}
	if !m.MatchRule("everyone", domain.MatchJID(domain.MustJID("bob", "localhost", ""))) {
		t.Fatal("bare allow must match all")
	}
}

func TestUserServiceAdminsFromEnv(t *testing.T) {
	cfg, err := Parse([]byte(`
auth_method: user_service
user_service:
  url: http://user-api-svc
  admins:
    - username: admin
      password: FromFile1
      email: ejabberd.admin@store.com
    - username: ops
      password: Ops12345
      email: ejabberd.ops@store.com
`))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("EJABBERD_ADMIN_USERNAME", "admin")
	t.Setenv("EJABBERD_ADMIN_PASSWORD", "FromSecret1")
	cfg.ApplyEnv()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(cfg.UserService.Admins) != 2 {
		t.Fatalf("admins = %+v", cfg.UserService.Admins)
	}
	for _, a := range cfg.UserService.Admins {
		if a.Username == "admin" && (a.Password != "FromSecret1" || a.Email != "ejabberd.admin@store.com") {
			t.Fatalf("env must override only the password: %+v", a)
		}
	}
	t.Setenv("EJABBERD_ADMIN_USERNAME", "root")
	t.Setenv("EJABBERD_ADMIN_PASSWORD", "")
	cfg.ApplyEnv()
	if err := cfg.Validate(); err == nil {
		t.Fatal("admin without password must be rejected")
	}
}
