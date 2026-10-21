package config

import (
	"strings"
	"testing"
)

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for _, k := range []string{"PORT", "OLLAMA_BASE_URL", "OLLAMA_MODEL", "MYSQL_DSN", "MYSQL_HOST", "MYSQL_PORT", "MYSQL_USER", "MYSQL_PASSWORD", "MYSQL_DATABASE"} {
		t.Setenv(k, env[k])
	}
}

func TestFromEnvPrefersDSN(t *testing.T) {
	setEnv(t, map[string]string{
		"OLLAMA_MODEL": "custom-model",
		"MYSQL_DSN":    "app:app@tcp(localhost:3307)/aiagent?parseTime=true",
		"MYSQL_HOST":   "ignored",
	})

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.MySQLDSN != "app:app@tcp(localhost:3307)/aiagent?parseTime=true" {
		t.Fatalf("dsn = %q", cfg.MySQLDSN)
	}
	if cfg.Port != "8080" || cfg.OllamaBaseURL != "http://localhost:11435" {
		t.Fatalf("defaults = %+v", cfg)
	}
}

func TestFromEnvBuildsDSNFromParts(t *testing.T) {
	setEnv(t, map[string]string{
		"OLLAMA_MODEL":   "custom-model",
		"MYSQL_HOST":     "mysql-svc",
		"MYSQL_PORT":     "80",
		"MYSQL_USER":     "ai_agent_system_svc",
		"MYSQL_PASSWORD": "p@ss:w/rd",
		"MYSQL_DATABASE": "ai_agent_system",
	})

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	want := "ai_agent_system_svc:p@ss:w/rd@tcp(mysql-svc:80)/ai_agent_system?parseTime=true"
	if cfg.MySQLDSN != want {
		t.Fatalf("dsn = %q, want %q", cfg.MySQLDSN, want)
	}
}

func TestFromEnvRequiredValues(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"model", map[string]string{"MYSQL_DSN": "x"}, "OLLAMA_MODEL"},
		{"mysql", map[string]string{"OLLAMA_MODEL": "m"}, "MYSQL_DSN or MYSQL_HOST"},
		{"database", map[string]string{"OLLAMA_MODEL": "m", "MYSQL_HOST": "h"}, "MYSQL_DATABASE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.env)
			if _, err := FromEnv(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want mention of %s", err, tc.want)
			}
		})
	}
}
