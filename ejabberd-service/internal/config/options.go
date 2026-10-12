package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"gopkg.in/yaml.v3"
)

func compilePattern(p string, glob bool) (*regexp.Regexp, error) {
	if glob {
		var b strings.Builder
		b.WriteByte('^')
		for _, r := range p {
			switch r {
			case '*':
				b.WriteString(".*")
			case '?':
				b.WriteByte('.')
			default:
				b.WriteString(regexp.QuoteMeta(string(r)))
			}
		}
		b.WriteByte('$')
		p = b.String()
	}
	return regexp.Compile(p)
}

func parseIPNet(s string) (*net.IPNet, error) {
	if !strings.Contains(s, "/") {
		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip %q", s)
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		s = fmt.Sprintf("%s/%d", s, bits)
	}
	_, n, err := net.ParseCIDR(s)
	return n, err
}

func (p *parser) modules(v *yaml.Node) error {
	m := &p.cfg.Modules
	return eachPair(v, func(name string, opts *yaml.Node) error {
		o := map[string]*yaml.Node{}
		if opts.Kind == yaml.MappingNode {
			_ = eachPair(opts, func(k string, val *yaml.Node) error {
				o[k] = val
				return nil
			})
		}
		str := func(k, def string) string {
			if n, ok := o[k]; ok {
				if n.Kind == yaml.ScalarNode {
					return n.Value
				}
				rule, err := parseAccessRule(n)
				if err == nil {
					inline := fmt.Sprintf("%s.%s", name, k)
					p.cfg.AccessRules[inline] = rule
					return inline
				}
			}
			return def
		}
		boolean := func(k string, def bool) bool {
			if n, ok := o[k]; ok {
				var b bool
				if n.Decode(&b) == nil {
					return b
				}
			}
			return def
		}
		dur := func(k string, def time.Duration) time.Duration {
			if n, ok := o[k]; ok {
				if d, err := parseTimeout(n.Value); err == nil {
					return d
				}
			}
			return def
		}
		switch name {
		case "mod_roster":
			m.Roster = RosterModule{Enabled: true, Versioning: boolean("versioning", false), Access: str("access", "all")}
		case "mod_offline":
			m.Offline = OfflineModule{
				Enabled:               true,
				AccessMaxUserMessages: str("access_max_user_messages", "max_user_offline_messages"),
				StoreGroupchat:        boolean("store_groupchat", false),
				StoreEmptyBody:        domain.StoreEmptyBody(str("store_empty_body", string(domain.StoreEmptyBodyUnlessChatState))),
			}
		case "mod_ping":
			interval := dur("ping_interval", time.Minute)
			m.Ping = PingModule{
				Enabled:        true,
				SendPings:      boolean("send_pings", false),
				PingInterval:   interval,
				PingAckTimeout: dur("ping_ack_timeout", interval),
				TimeoutAction:  str("timeout_action", "none"),
			}
		case "mod_disco":
			m.Disco = true
		case "mod_version":
			m.Version = VersionModule{Enabled: true, ShowOS: boolean("show_os", true)}
		case "mod_register":
			strength := 0
			if n, ok := o["password_strength"]; ok {
				strength, _ = strconv.Atoi(n.Value)
			}
			var watch []string
			if n, ok := o["registration_watchers"]; ok {
				_ = n.Decode(&watch)
			}
			m.Register = RegisterModule{
				Enabled:           true,
				Access:            str("access", "all"),
				IPAccess:          str("ip_access", "all"),
				AccessRemove:      str("access_remove", "all"),
				AccessFrom:        str("access_from", "none"),
				PasswordStrength:  strength,
				RedirectURL:       str("redirect_url", ""),
				RegistrationWatch: watch,
			}
		case "mod_http_api":
			m.HTTPAPI = true
		case "mod_http_upload":
			var maxSize int64 = 104857600
			if n, ok := o["max_size"]; ok {
				if n.Value == "infinity" {
					maxSize = 0
				} else if err := n.Decode(&maxSize); err != nil {
					return fmt.Errorf("mod_http_upload max_size: %w", err)
				}
			}
			m.HTTPUpload = HTTPUploadModule{
				Enabled:        true,
				Host:           str("host", "upload.@HOST@"),
				Name:           str("name", "HTTP File Upload"),
				Access:         str("access", "local"),
				MaxSize:        maxSize,
				PutURL:         str("put_url", "https://@HOST@:5443/upload"),
				GetURL:         str("get_url", ""),
				ExternalSecret: str("external_secret", ""),
			}
		case "mod_last":
			m.Last = true
		default:
			p.cfg.Unsupported = append(p.cfg.Unsupported, "modules."+name)
		}
		return nil
	})
}

func (p *parser) userService(v *yaml.Node) error {
	u := &p.cfg.UserService
	return eachPair(v, func(k string, n *yaml.Node) error {
		switch k {
		case "url":
			u.URL = n.Value
		case "internal_token":
			u.InternalToken = n.Value
		case "jid_field":
			u.JIDField = n.Value
		case "allow_password_login":
			return n.Decode(&u.AllowPasswordLogin)
		case "timeout":
			d, err := parseTimeout(n.Value)
			if err != nil {
				return err
			}
			u.Timeout = d
		case "admins":
			var raw []struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
				Email    string `yaml:"email"`
				Name     string `yaml:"name"`
			}
			if err := n.Decode(&raw); err != nil {
				return fmt.Errorf("user_service.admins: %w", err)
			}
			for _, a := range raw {
				u.Admins = append(u.Admins, UserServiceAccount(a))
			}
		case "cache_ttl":
			d, err := parseTimeout(n.Value)
			if err != nil {
				return err
			}
			u.CacheTTL = d
		default:
			return fmt.Errorf("user_service: unknown option %q", k)
		}
		return nil
	})
}

func (p *parser) apiPermissions(v *yaml.Node) error {
	return eachPair(v, func(name string, val *yaml.Node) error {
		perm := APIPermission{Name: name}
		err := eachPair(val, func(k string, n *yaml.Node) error {
			switch k {
			case "from":
				perm.From = scalarList(n)
			case "what":
				perm.What = scalarList(n)
			case "who":
				who, err := parseWho(n)
				if err != nil {
					return err
				}
				perm.Who = who
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("api_permissions %s: %w", name, err)
		}
		p.cfg.APIPermissions = append(p.cfg.APIPermissions, perm)
		return nil
	})
}

func parseWho(n *yaml.Node) ([]WhoClause, error) {
	switch n.Kind {
	case yaml.ScalarNode:
		if n.Value == "all" {
			return []WhoClause{{All: true}}, nil
		}
		return []WhoClause{{Specs: []domain.ACLSpec{{Kind: domain.ACLRef, Name: n.Value}}}}, nil
	case yaml.SequenceNode:
		var out []WhoClause
		for _, item := range n.Content {
			c, err := parseWho(item)
			if err != nil {
				return nil, err
			}
			out = append(out, c...)
		}
		return out, nil
	case yaml.MappingNode:
		var out []WhoClause
		err := eachPair(n, func(k string, val *yaml.Node) error {
			switch k {
			case "access":
				rule, err := parseAccessRule(val)
				if err != nil {
					return err
				}
				out = append(out, WhoClause{Access: rule})
			case "oauth":
			case "acl":
				out = append(out, WhoClause{Specs: []domain.ACLSpec{{Kind: domain.ACLRef, Name: val.Value}}})
			default:
				spec, err := parseACLSpec(domain.ACLKind(k), val, false)
				if err != nil {
					return err
				}
				out = append(out, WhoClause{Specs: []domain.ACLSpec{spec}})
			}
			return nil
		})
		return out, err
	}
	return nil, errors.New("invalid who")
}

func scalarList(n *yaml.Node) []string {
	if n.Kind == yaml.ScalarNode {
		return []string{n.Value}
	}
	var out []string
	for _, c := range n.Content {
		out = append(out, c.Value)
	}
	return out
}

func (c *Config) ApplyEnv() {
	if v := os.Getenv("EJABBERD_HOSTS"); v != "" {
		c.Hosts = splitList(v)
	}
	setString(&c.LogLevel, "EJABBERD_LOGLEVEL")
	setString(&c.DefaultDB, "EJABBERD_DEFAULT_DB")
	setString(&c.AuthMethod, "EJABBERD_AUTH_METHOD")
	setString(&c.UserService.URL, "USER_SERVICE_URL")
	setString(&c.UserService.InternalToken, "USER_SERVICE_INTERNAL_TOKEN")
	if name := os.Getenv("EJABBERD_ADMIN_USERNAME"); name != "" {
		admin := UserServiceAccount{Username: name}
		kept := c.UserService.Admins[:0]
		for _, a := range c.UserService.Admins {
			if a.Username == name {
				admin = a
				continue
			}
			kept = append(kept, a)
		}
		setString(&admin.Password, "EJABBERD_ADMIN_PASSWORD")
		setString(&admin.Email, "EJABBERD_ADMIN_EMAIL")
		c.UserService.Admins = append(kept, admin)
	}
	setString(&c.UploadService.URL, "UPLOAD_SERVICE_URL")
	setString(&c.UploadService.ServiceKey, "UPLOAD_SERVICE_KEY")
	setString(&c.Modules.HTTPUpload.PutURL, "HTTP_UPLOAD_PUT_URL")
	setString(&c.Modules.HTTPUpload.GetURL, "HTTP_UPLOAD_GET_URL")
	setString(&c.Modules.HTTPUpload.ExternalSecret, "HTTP_UPLOAD_SECRET")
	setString(&c.SQL.Server, "SQL_SERVER")
	setString(&c.SQL.Database, "SQL_DATABASE")
	setString(&c.SQL.Username, "SQL_USERNAME")
	setString(&c.SQL.Password, "SQL_PASSWORD")
	if v, err := strconv.Atoi(os.Getenv("SQL_PORT")); err == nil {
		c.SQL.Port = v
	}
	if v := os.Getenv("EJABBERD_CERTFILES"); v != "" {
		c.CertFiles = splitList(v)
	}
}

func (c *Config) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("config hosts is empty")
	}
	for i, h := range c.Hosts {
		lh, ok := domain.Nameprep(h)
		if !ok {
			return fmt.Errorf("config hosts: invalid host %q", h)
		}
		c.Hosts[i] = lh
	}
	switch c.AuthPasswordFormat {
	case domain.PasswordPlain, domain.PasswordScram:
	default:
		return fmt.Errorf("config auth_password_format must be plain or scram, got %q", c.AuthPasswordFormat)
	}
	switch c.AuthScramHash {
	case domain.ScramSHA1, domain.ScramSHA256, domain.ScramSHA512:
	default:
		return fmt.Errorf("config auth_scram_hash must be sha, sha256 or sha512, got %q", c.AuthScramHash)
	}
	switch c.ResourceConflict {
	case domain.ConflictAcceptNew, domain.ConflictCloseOld, domain.ConflictCloseNew, domain.ConflictSetResource:
	default:
		return fmt.Errorf("config resource_conflict invalid: %q", c.ResourceConflict)
	}
	switch c.DefaultDB {
	case "mnesia", "sql":
	default:
		return fmt.Errorf("config default_db must be mnesia or sql, got %q", c.DefaultDB)
	}
	switch c.AuthMethod {
	case "internal", "sql":
	case "user_service":
		if c.UserService.URL == "" {
			return errors.New("config auth_method user_service requires user_service.url (or USER_SERVICE_URL)")
		}
		switch c.UserService.JIDField {
		case "username", "user_id":
		default:
			return fmt.Errorf("config user_service.jid_field must be username or user_id, got %q", c.UserService.JIDField)
		}
		for _, a := range c.UserService.Admins {
			if a.Username == "" || a.Password == "" {
				return fmt.Errorf("config user_service.admins: username and password are required (admin %q)", a.Username)
			}
		}
	default:
		return fmt.Errorf("config auth_method must be internal, sql or user_service, got %q", c.AuthMethod)
	}
	if c.Modules.HTTPUpload.Enabled && (c.UploadService.URL == "" || c.UploadService.ServiceKey == "") {
		return errors.New("config mod_http_upload stores files in upload_service: set upload_service.url and upload_service.service_key (or UPLOAD_SERVICE_URL / UPLOAD_SERVICE_KEY)")
	}
	if c.UseSQL() && c.SQL.Type != "pgsql" {
		return fmt.Errorf("config sql_type %q is not supported, only pgsql", c.SQL.Type)
	}
	return nil
}

func ExpandHost(s, host string) string {
	return strings.ReplaceAll(s, "@HOST@", host)
}

func (c *Config) Matcher() *domain.ACLMatcher {
	return &domain.ACLMatcher{
		ACLs:        c.ACL,
		AccessRules: c.AccessRules,
		ShaperRules: c.ShaperRules,
		IsMyHost:    c.IsMyHost,
	}
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		*dst = v
	}
}

func splitList(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
