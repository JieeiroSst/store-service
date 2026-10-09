package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var Version = "dev"

type Config struct {
	Server        ServerConfig        `json:"server"`
	Secret        SecretConfig        `json:"secret"`
	Elasticsearch ElasticsearchConfig `json:"elasticsearch"`
}

type ServerConfig struct {
	ServerPort string `json:"server_port"`
}

type SecretConfig struct {
	AuthorizeKey string `json:"authorize_key"`
}

type ElasticsearchConfig struct {
	DNS              string       `json:"dns"`
	Username         string       `json:"username"`
	Password         string       `json:"password"`
	Indices          []string     `json:"indices"`
	SensitivePattern string       `json:"sensitive_pattern"`
	Schema           SchemaConfig `json:"schema"`
}

type SchemaConfig struct {
	Name             string   `json:"name"`
	IndexPatterns    []string `json:"index_patterns"`
	Priority         int      `json:"priority"`
	NumberOfShards   int      `json:"number_of_shards"`
	NumberOfReplicas *int     `json:"number_of_replicas"`
	Synonyms         []string `json:"synonyms"`
	ApplyOnStartup   *bool    `json:"apply_on_startup"`
}

func (s SchemaConfig) ShouldApplyOnStartup() bool {
	return s.ApplyOnStartup == nil || *s.ApplyOnStartup
}

func Defaults() *Config {
	return &Config{
		Server:        ServerConfig{ServerPort: "8080"},
		Elasticsearch: ElasticsearchConfig{DNS: "http://localhost:9200"},
	}
}

func (c *Config) Merge(raw []byte) error {
	if err := json.Unmarshal(raw, c); err != nil {
		return fmt.Errorf("parse config json: %w", err)
	}
	return nil
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.ServerPort, "PORT_HTTP_SERVER")
	setString(&c.Secret.AuthorizeKey, "AUTHORIZE_KEY")
	setString(&c.Elasticsearch.DNS, "ELASTICSEARCH_URL")
	setString(&c.Elasticsearch.Username, "ELASTICSEARCH_USERNAME")
	setString(&c.Elasticsearch.Password, "ELASTICSEARCH_PASSWORD")
	if v, ok := os.LookupEnv("ELASTICSEARCH_INDICES"); ok {
		c.Elasticsearch.Indices = splitList(v)
	}
	if v, err := strconv.ParseBool(os.Getenv("SCHEMA_APPLY_ON_STARTUP")); err == nil {
		c.Elasticsearch.Schema.ApplyOnStartup = &v
	}
}

func (c *Config) Validate() error {
	if c.Server.ServerPort == "" {
		return fmt.Errorf("config server.server_port is empty")
	}
	if c.Secret.AuthorizeKey == "" {
		return fmt.Errorf("config secret.authorize_key (or AUTHORIZE_KEY) is required")
	}
	c.Elasticsearch.DNS = strings.TrimRight(c.Elasticsearch.DNS, "/")
	if u, err := url.Parse(c.Elasticsearch.DNS); err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("config elasticsearch.dns must be an absolute URL, got %q", c.Elasticsearch.DNS)
	}
	if c.Elasticsearch.Schema.Priority < 0 || c.Elasticsearch.Schema.NumberOfShards < 0 {
		return fmt.Errorf("config elasticsearch.schema.priority and number_of_shards must not be negative")
	}
	if r := c.Elasticsearch.Schema.NumberOfReplicas; r != nil && *r < 0 {
		return fmt.Errorf("config elasticsearch.schema.number_of_replicas must not be negative")
	}
	return nil
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
