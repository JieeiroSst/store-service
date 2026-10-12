package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"gopkg.in/yaml.v3"
)

var Version = "dev"

type Listener struct {
	Port             int
	IP               string
	Module           string
	Transport        string
	MaxStanzaSize    int
	Shaper           string
	Access           string
	StartTLS         bool
	StartTLSRequired bool
	TLS              bool
	RequestHandlers  map[string]string
}

type SQLConfig struct {
	Type     string
	Server   string
	Port     int
	Database string
	Username string
	Password string
	SSL      bool
	PoolSize int
}

func (s SQLConfig) DSN() string {
	ssl := "disable"
	if s.SSL {
		ssl = "require"
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		s.Server, s.Port, s.Database, s.Username, s.Password, ssl)
}

type UserServiceAccount struct {
	Username string
	Password string
	Email    string
	Name     string
}

type UserServiceConfig struct {
	Admins             []UserServiceAccount
	URL                string
	Timeout            time.Duration
	InternalToken      string
	JIDField           string
	AllowPasswordLogin bool
	CacheTTL           time.Duration
}

type UploadServiceConfig struct {
	URL        string
	ServiceKey string
	Timeout    time.Duration
}

type HTTPUploadModule struct {
	Enabled        bool
	Host           string
	Name           string
	Access         string
	MaxSize        int64
	PutURL         string
	GetURL         string
	ExternalSecret string
}

type RosterModule struct {
	Enabled    bool
	Versioning bool
	Access     string
}

type OfflineModule struct {
	Enabled               bool
	AccessMaxUserMessages string
	StoreGroupchat        bool
	StoreEmptyBody        domain.StoreEmptyBody
}

type PingModule struct {
	Enabled        bool
	SendPings      bool
	PingInterval   time.Duration
	PingAckTimeout time.Duration
	TimeoutAction  string
}

type VersionModule struct {
	Enabled bool
	ShowOS  bool
}

type RegisterModule struct {
	Enabled           bool
	Access            string
	IPAccess          string
	AccessRemove      string
	AccessFrom        string
	PasswordStrength  int
	RedirectURL       string
	RegistrationWatch []string
}

type Modules struct {
	Roster     RosterModule
	Offline    OfflineModule
	Ping       PingModule
	Disco      bool
	Version    VersionModule
	Register   RegisterModule
	HTTPAPI    bool
	HTTPUpload HTTPUploadModule
	Last       bool
}

type APIPermission struct {
	Name string
	From []string
	Who  []WhoClause
	What []string
}

type WhoClause struct {
	All    bool
	Access domain.AccessRule
	Specs  []domain.ACLSpec
}

type Config struct {
	Hosts               []string
	LogLevel            string
	Language            string
	CertFiles           []string
	Listen              []Listener
	ACL                 map[string][]domain.ACLSpec
	AccessRules         map[string]domain.AccessRule
	Shapers             map[string]domain.ShaperSpec
	ShaperRules         map[string]domain.ShaperRule
	APIPermissions      []APIPermission
	AuthMethod          string
	DefaultDB           string
	AuthPasswordFormat  domain.PasswordFormat
	AuthScramHash       domain.ScramHash
	ResourceConflict    domain.ResourceConflict
	NegotiationTimeout  time.Duration
	RegistrationTimeout time.Duration
	SQL                 SQLConfig
	UserService         UserServiceConfig
	UploadService       UploadServiceConfig
	Modules             Modules
	Unsupported         []string
}

func (c *Config) IsMyHost(host string) bool {
	return slices.Contains(c.Hosts, host)
}

func (c *Config) MyName() string {
	if len(c.Hosts) == 0 {
		return "localhost"
	}
	return c.Hosts[0]
}

func (c *Config) ExternalAuth() bool {
	return c.AuthMethod == "user_service"
}

func (c *Config) UseSQL() bool {
	return c.DefaultDB == "sql" || c.AuthMethod == "sql"
}

func Load() (*Config, error) {
	path := os.Getenv("EJABBERD_CONFIG")
	if path == "" {
		for _, p := range []string{"/etc/ejabberd/ejabberd.yml", "ejabberd.yml"} {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	var raw []byte
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}
		raw = b
		slog.Info("loading config", "path", path)
	}
	cfg, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	cfg.ApplyEnv()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Defaults() *Config {
	return &Config{
		Hosts:               []string{"localhost"},
		LogLevel:            "info",
		Language:            "en",
		ACL:                 map[string][]domain.ACLSpec{},
		AccessRules:         map[string]domain.AccessRule{},
		Shapers:             map[string]domain.ShaperSpec{},
		ShaperRules:         map[string]domain.ShaperRule{},
		AuthMethod:          "internal",
		DefaultDB:           "mnesia",
		AuthPasswordFormat:  domain.PasswordPlain,
		AuthScramHash:       domain.ScramSHA1,
		ResourceConflict:    domain.ConflictAcceptNew,
		NegotiationTimeout:  120 * time.Second,
		RegistrationTimeout: 600 * time.Second,
		UserService: UserServiceConfig{
			Timeout:            5 * time.Second,
			JIDField:           "username",
			AllowPasswordLogin: true,
			CacheTTL:           5 * time.Minute,
		},
		UploadService: UploadServiceConfig{Timeout: 60 * time.Second},
		SQL:           SQLConfig{Type: "pgsql", Server: "localhost", Port: 5432, Database: "ejabberd", Username: "ejabberd", PoolSize: 10},
	}
}

func Parse(raw []byte) (*Config, error) {
	cfg := Defaults()
	if len(strings.TrimSpace(string(raw))) == 0 {
		return cfg, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}
	if len(doc.Content) == 0 {
		return cfg, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, errors.New("config root must be a mapping")
	}
	p := &parser{cfg: cfg}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if err := p.option(root.Content[i].Value, root.Content[i+1]); err != nil {
			return nil, fmt.Errorf("option %s: %w", root.Content[i].Value, err)
		}
	}
	return cfg, nil
}

type parser struct {
	cfg *Config
}

func (p *parser) option(key string, v *yaml.Node) error {
	c := p.cfg
	switch key {
	case "hosts":
		return v.Decode(&c.Hosts)
	case "loglevel":
		c.LogLevel = v.Value
	case "language":
		c.Language = v.Value
	case "certfiles":
		return v.Decode(&c.CertFiles)
	case "listen":
		return p.listen(v)
	case "acl":
		return p.acl(v)
	case "access_rules":
		return p.accessRules(v)
	case "shaper":
		return p.shapers(v)
	case "shaper_rules":
		return p.shaperRules(v)
	case "api_permissions":
		return p.apiPermissions(v)
	case "auth_method":
		c.AuthMethod = firstScalar(v)
	case "default_db":
		c.DefaultDB = v.Value
	case "auth_password_format":
		c.AuthPasswordFormat = domain.PasswordFormat(v.Value)
	case "auth_scram_hash":
		c.AuthScramHash = domain.ScramHash(v.Value)
	case "resource_conflict":
		c.ResourceConflict = domain.ResourceConflict(v.Value)
	case "negotiation_timeout":
		d, err := parseTimeout(v.Value)
		if err != nil {
			return err
		}
		c.NegotiationTimeout = d
	case "registration_timeout":
		if v.Value == "infinity" {
			c.RegistrationTimeout = 0
			return nil
		}
		d, err := parseTimeout(v.Value)
		if err != nil {
			return err
		}
		c.RegistrationTimeout = d
	case "user_service":
		return p.userService(v)
	case "upload_service":
		return eachPair(v, func(k string, n *yaml.Node) error {
			switch k {
			case "url":
				c.UploadService.URL = n.Value
			case "service_key":
				c.UploadService.ServiceKey = n.Value
			case "timeout":
				d, err := parseTimeout(n.Value)
				if err != nil {
					return err
				}
				c.UploadService.Timeout = d
			default:
				return fmt.Errorf("upload_service: unknown option %q", k)
			}
			return nil
		})
	case "sql_type":
		c.SQL.Type = v.Value
	case "sql_server":
		c.SQL.Server = v.Value
	case "sql_port":
		return v.Decode(&c.SQL.Port)
	case "sql_database":
		c.SQL.Database = v.Value
	case "sql_username":
		c.SQL.Username = v.Value
	case "sql_password":
		c.SQL.Password = v.Value
	case "sql_ssl":
		return v.Decode(&c.SQL.SSL)
	case "sql_pool_size":
		return v.Decode(&c.SQL.PoolSize)
	case "new_sql_schema", "sql_schema_update", "update_sql_schema":
	case "modules":
		return p.modules(v)
	default:
		c.Unsupported = append(c.Unsupported, key)
	}
	return nil
}

func firstScalar(v *yaml.Node) string {
	if v.Kind == yaml.SequenceNode && len(v.Content) > 0 {
		return v.Content[0].Value
	}
	return v.Value
}

func parseTimeout(s string) (time.Duration, error) {
	if n, err := strconv.Atoi(s); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	fields := strings.Fields(s)
	if len(fields) == 2 {
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			return 0, err
		}
		switch strings.TrimSuffix(fields[1], "s") {
		case "sec", "second":
			return time.Duration(n) * time.Second, nil
		case "min", "minute":
			return time.Duration(n) * time.Minute, nil
		case "hour":
			return time.Duration(n) * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}

func (p *parser) listen(v *yaml.Node) error {
	if v.Kind != yaml.SequenceNode {
		return errors.New("must be a list")
	}
	for _, item := range v.Content {
		l := Listener{IP: "::", Transport: "tcp", Access: "all", Shaper: "none"}
		var raw map[string]yaml.Node
		if err := item.Decode(&raw); err != nil {
			return err
		}
		for k, n := range raw {
			var err error
			switch k {
			case "port":
				err = n.Decode(&l.Port)
			case "ip":
				l.IP = n.Value
			case "module":
				l.Module = n.Value
			case "transport":
				l.Transport = n.Value
			case "max_stanza_size":
				if n.Value != "infinity" {
					err = n.Decode(&l.MaxStanzaSize)
				}
			case "shaper":
				l.Shaper = n.Value
			case "access":
				l.Access = n.Value
			case "starttls":
				err = n.Decode(&l.StartTLS)
			case "starttls_required":
				err = n.Decode(&l.StartTLSRequired)
			case "tls":
				err = n.Decode(&l.TLS)
			case "request_handlers":
				err = n.Decode(&l.RequestHandlers)
			}
			if err != nil {
				return fmt.Errorf("listener %s: %w", k, err)
			}
		}
		if l.StartTLSRequired {
			l.StartTLS = true
		}
		p.cfg.Listen = append(p.cfg.Listen, l)
	}
	return nil
}

func (p *parser) acl(v *yaml.Node) error {
	return eachPair(v, func(name string, val *yaml.Node) error {
		specs, err := parseACLSpecs(val, false)
		if err != nil {
			return fmt.Errorf("acl %s: %w", name, err)
		}
		p.cfg.ACL[name] = specs
		return nil
	})
}

func (p *parser) accessRules(v *yaml.Node) error {
	return eachPair(v, func(name string, val *yaml.Node) error {
		rule, err := parseAccessRule(val)
		if err != nil {
			return fmt.Errorf("access rule %s: %w", name, err)
		}
		p.cfg.AccessRules[name] = rule
		return nil
	})
}

func (p *parser) shapers(v *yaml.Node) error {
	return eachPair(v, func(name string, val *yaml.Node) error {
		var spec domain.ShaperSpec
		switch val.Kind {
		case yaml.ScalarNode:
			if err := val.Decode(&spec.Rate); err != nil {
				return err
			}
		case yaml.MappingNode:
			var m struct {
				Rate      int `yaml:"rate"`
				BurstSize int `yaml:"burst_size"`
			}
			if err := val.Decode(&m); err != nil {
				return err
			}
			spec = domain.ShaperSpec{Rate: m.Rate, Burst: m.BurstSize}
		}
		p.cfg.Shapers[name] = spec
		return nil
	})
}

func (p *parser) shaperRules(v *yaml.Node) error {
	return eachPair(v, func(name string, val *yaml.Node) error {
		rule, err := parseShaperRule(val)
		if err != nil {
			return fmt.Errorf("shaper rule %s: %w", name, err)
		}
		p.cfg.ShaperRules[name] = rule
		return nil
	})
}

func parseShaperRule(v *yaml.Node) (domain.ShaperRule, error) {
	all := []domain.ACLSpec{{Kind: domain.ACLRef, Name: "all"}}
	switch v.Kind {
	case yaml.ScalarNode:
		return domain.ShaperRule{{Value: v.Value, Specs: all}}, nil
	case yaml.MappingNode:
		var rule domain.ShaperRule
		err := eachPair(v, func(value string, who *yaml.Node) error {
			specs, err := parseAccessSpecs(who)
			if err != nil {
				return err
			}
			rule = append(rule, domain.ShaperClause{Value: value, Specs: specs})
			return nil
		})
		return rule, err
	case yaml.SequenceNode:
		var rule domain.ShaperRule
		for _, item := range v.Content {
			switch item.Kind {
			case yaml.ScalarNode:
				rule = append(rule, domain.ShaperClause{Value: item.Value, Specs: all})
			case yaml.MappingNode:
				sub, err := parseShaperRule(item)
				if err != nil {
					return nil, err
				}
				rule = append(rule, sub...)
			}
		}
		return rule, nil
	}
	return nil, errors.New("invalid shaper rule")
}

func parseAccessRule(v *yaml.Node) (domain.AccessRule, error) {
	var rule domain.AccessRule
	add := func(action string, who *yaml.Node) error {
		var allow bool
		switch action {
		case "allow":
			allow = true
		case "deny":
		default:
			return fmt.Errorf("unknown action %q", action)
		}
		specs := []domain.ACLSpec{{Kind: domain.ACLRef, Name: "all"}}
		if who != nil {
			var err error
			if specs, err = parseAccessSpecs(who); err != nil {
				return err
			}
		}
		rule = append(rule, domain.AccessClause{Allow: allow, Specs: specs})
		return nil
	}
	switch v.Kind {
	case yaml.MappingNode:
		err := eachPair(v, func(k string, val *yaml.Node) error { return add(k, val) })
		return rule, err
	case yaml.SequenceNode:
		for _, item := range v.Content {
			switch item.Kind {
			case yaml.ScalarNode:
				if err := add(item.Value, nil); err != nil {
					return nil, err
				}
			case yaml.MappingNode:
				if err := eachPair(item, func(k string, val *yaml.Node) error { return add(k, val) }); err != nil {
					return nil, err
				}
			}
		}
		return rule, nil
	case yaml.ScalarNode:
		return rule, add(v.Value, nil)
	}
	return nil, errors.New("invalid access rule")
}

func parseAccessSpecs(v *yaml.Node) ([]domain.ACLSpec, error) {
	switch v.Kind {
	case yaml.ScalarNode:
		return []domain.ACLSpec{{Kind: domain.ACLRef, Name: v.Value}}, nil
	case yaml.SequenceNode:
		var out []domain.ACLSpec
		for _, item := range v.Content {
			specs, err := parseAccessSpecs(item)
			if err != nil {
				return nil, err
			}
			out = append(out, specs...)
		}
		return out, nil
	case yaml.MappingNode:
		return parseACLSpecs(v, true)
	}
	return nil, errors.New("invalid access specification")
}

func parseACLSpecs(v *yaml.Node, allowRef bool) ([]domain.ACLSpec, error) {
	if v.Kind == yaml.SequenceNode {
		var out []domain.ACLSpec
		for _, item := range v.Content {
			specs, err := parseACLSpecs(item, allowRef)
			if err != nil {
				return nil, err
			}
			out = append(out, specs...)
		}
		return out, nil
	}
	var out []domain.ACLSpec
	err := eachPair(v, func(kind string, val *yaml.Node) error {
		values := []*yaml.Node{val}
		if val.Kind == yaml.SequenceNode {
			values = val.Content
		}
		for _, item := range values {
			spec, err := parseACLSpec(domain.ACLKind(kind), item, allowRef)
			if err != nil {
				return err
			}
			out = append(out, spec)
		}
		return nil
	})
	return out, err
}

func parseACLSpec(kind domain.ACLKind, v *yaml.Node, allowRef bool) (domain.ACLSpec, error) {
	spec := domain.ACLSpec{Kind: kind}
	switch kind {
	case domain.ACLRef:
		if !allowRef {
			return spec, errors.New("acl reference not allowed here")
		}
		spec.Name = v.Value
	case domain.ACLUser:
		if v.Kind == yaml.MappingNode {
			return spec, eachPair(v, func(u string, s *yaml.Node) error {
				spec.User, spec.Server = u, s.Value
				return nil
			})
		}
		u, s, _ := strings.Cut(v.Value, "@")
		lu, _ := domain.Nodeprep(u)
		spec.User = lu
		if s != "" {
			spec.Server, _ = domain.Nameprep(s)
		}
	case domain.ACLServer:
		spec.Server, _ = domain.Nameprep(v.Value)
	case domain.ACLResource:
		spec.Resource = v.Value
	case domain.ACLUserRegexp, domain.ACLResourceRegexp, domain.ACLUserGlob, domain.ACLResourceGlob:
		pattern, server := v.Value, ""
		if v.Kind == yaml.MappingNode {
			if err := eachPair(v, func(u string, s *yaml.Node) error {
				pattern, server = u, s.Value
				return nil
			}); err != nil {
				return spec, err
			}
		} else if kind == domain.ACLUserRegexp || kind == domain.ACLUserGlob {
			if u, s, ok := strings.Cut(pattern, "@"); ok {
				pattern, server = u, s
			}
		}
		re, err := compilePattern(pattern, kind == domain.ACLUserGlob || kind == domain.ACLResourceGlob)
		if err != nil {
			return spec, err
		}
		spec.UserRE, spec.Server = re, server
	case domain.ACLServerRegexp, domain.ACLServerGlob:
		re, err := compilePattern(v.Value, kind == domain.ACLServerGlob)
		if err != nil {
			return spec, err
		}
		spec.ServerRE = re
	case domain.ACLNodeRegexp, domain.ACLNodeGlob:
		glob := kind == domain.ACLNodeGlob
		return spec, eachPair(v, func(u string, s *yaml.Node) error {
			var err error
			if spec.UserRE, err = compilePattern(u, glob); err != nil {
				return err
			}
			spec.ServerRE, err = compilePattern(s.Value, glob)
			return err
		})
	case domain.ACLIP:
		n, err := parseIPNet(v.Value)
		if err != nil {
			return spec, err
		}
		spec.Net = n
	default:
		return spec, fmt.Errorf("unsupported acl type %q", kind)
	}
	return spec, nil
}

func eachPair(v *yaml.Node, fn func(string, *yaml.Node) error) error {
	if v.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping at line %d", v.Line)
	}
	for i := 0; i+1 < len(v.Content); i += 2 {
		if err := fn(v.Content[i].Value, v.Content[i+1]); err != nil {
			return err
		}
	}
	return nil
}
