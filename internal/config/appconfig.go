package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	VerifyTLS *bool  `yaml:"verify_tls"`
}

type Server struct {
	Listen         string `yaml:"listen"`
	DefaultRefresh int    `yaml:"default_refresh"`
}

type OIDCConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Issuer          string   `yaml:"issuer"`
	ClientID        string   `yaml:"client_id"`
	ClientSecret    string   `yaml:"client_secret"`
	SessionTTLHours int      `yaml:"session_ttl_hours"`
	RedirectURI     string   `yaml:"redirect_uri"`
	AllowedGroups   []string `yaml:"allowed_groups"`
}

func (o *OIDCConfig) EnabledAndConfigured() bool {
	return o != nil && o.Enabled && o.Issuer != "" && o.ClientID != ""
}

func (o *OIDCConfig) SessionTTL() time.Duration {
	ttl := time.Duration(o.SessionTTLHours) * time.Hour
	if ttl <= 0 {
		return 8 * time.Hour
	}
	return ttl
}

type AppConfig struct {
	Server       Server     `yaml:"server"`
	Targets      []Target   `yaml:"targets"`
	RequireAuth  bool       `yaml:"require_auth"`
	TokensFile   string     `yaml:"tokens_file"`
	TokenTTLDays int        `yaml:"token_ttl_days"`
	LogPath      string     `yaml:"log_path"`
	OIDC         OIDCConfig `yaml:"oidc"`
}

func Load(path string) (*AppConfig, error) {
	if path == "" {
		path = os.Getenv("PROXYGLASS_CONFIG")
	}
	if path == "" {
		path = "/etc/proxyglass/config.yml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		// fallback to local
		if path != "config.yaml" && path != "./config.yaml" {
			if data2, err2 := os.ReadFile("config.yaml"); err2 == nil {
				data = data2
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	var c AppConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	var refreshConfig struct {
		Server struct {
			DefaultRefresh *int `yaml:"default_refresh"`
		} `yaml:"server"`
	}
	if err := yaml.Unmarshal(data, &refreshConfig); err != nil {
		return nil, err
	}
	if c.Server.Listen == "" {
		c.Server.Listen = ":8080"
	}
	if refreshConfig.Server.DefaultRefresh == nil {
		c.Server.DefaultRefresh = 5
	}
	if c.TokensFile == "" {
		c.TokensFile = "/etc/proxyglass/tokens.json"
	}
	return &c, nil
}

func GetDefaultConfigYML() string {
	return `server:
  listen: "127.0.0.1:8080"
  default_refresh: 5

# Require authentication for dashboard access (supports Bearer tokens and Basic Auth)
require_auth: true

tokens_file: "/etc/proxyglass/tokens.json"
token_ttl_days: 0
log_path: "/var/log/proxyglass/proxyglass.log"

# Optional OIDC SSO for dashboard login
oidc:
  enabled: false
  issuer: ""
  client_id: ""
  client_secret: ""
  session_ttl_hours: 8
  redirect_uri: ""
  allowed_groups: []

targets:
  - name: haproxy-lb-01
    url: http://127.0.0.1:8404/stats;csv
    username: admin
    password: changeme
    verify_tls: false
`
}
