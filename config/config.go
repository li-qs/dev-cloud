package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

const (
	defaultServerAddr   = ":8080"
	defaultAccessTTL    = 900
	defaultRefreshTTL   = 604800
	defaultCookieSecure = true
)

type Config struct {
	ServerAddr   string   `yaml:"server_addr"`
	Postgres     string   `yaml:"postgres"`
	JWTSecret    string   `yaml:"jwt_secret"`
	TokenSalt    string   `yaml:"token_salt"`
	AccessTTL    int      `yaml:"access_ttl"`
	RefreshTTL   int      `yaml:"refresh_ttl"`
	CookieSecure *bool    `yaml:"cookie_secure"`
}

func LoadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = defaultAccessTTL
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = defaultRefreshTTL
	}
	if cfg.CookieSecure == nil {
		secure := defaultCookieSecure
		cfg.CookieSecure = &secure
	}

	if cfg.Postgres == "" {
		return fmt.Errorf("postgres is required")
	}
	if cfg.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}

	return nil
}
