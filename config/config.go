package config

import (
	"fmt"
	"os"
	"runtime"

	"go.yaml.in/yaml/v4"
)

const (
	defaultServerAddr   = ":8080"
	defaultAccessTTL    = 900
	defaultRefreshTTL   = 604800
	defaultCookieSecure = true
)

type Config struct {
	ServerAddr   string       `yaml:"server_addr"`
	Postgres     string       `yaml:"postgres"`
	Redis        RedisConfig  `yaml:"redis"`
	JWTSecret    string       `yaml:"jwt_secret"`
	TokenSalt    string       `yaml:"token_salt"`
	AccessTTL    int          `yaml:"access_ttl"`
	RefreshTTL   int          `yaml:"refresh_ttl"`
	CookieSecure *bool        `yaml:"cookie_secure"`
	WorkerPool   int          `yaml:"worker_pool"`
	Docker       DockerConfig `yaml:"docker"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type DockerConfig struct {
	Host string `yaml:"host"`
	TLS  struct {
		Enabled bool   `yaml:"enabled"`
		CA      string `yaml:"ca"`
		Cert    string `yaml:"cert"`
		Key     string `yaml:"key"`
	} `yaml:"tls"`
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
	if cfg.WorkerPool == 0 {
		cfg.WorkerPool = runtime.NumCPU() * 2
	}

	// 敏感配置优先从环境变量读取，避免密钥写入配置文件/仓库。
	cfg.Postgres = envOr("POSTGRES_DSN", cfg.Postgres)
	cfg.JWTSecret = envOr("JWT_SECRET", cfg.JWTSecret)
	cfg.TokenSalt = envOr("TOKEN_SALT", cfg.TokenSalt)
	cfg.Redis.Password = envOr("REDIS_PASSWORD", cfg.Redis.Password)

	if cfg.Postgres == "" {
		return fmt.Errorf("postgres is required")
	}
	if cfg.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if cfg.TokenSalt == "" {
		return fmt.Errorf("token_salt is required")
	}

	return nil
}

func envOr(key, current string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return current
}
