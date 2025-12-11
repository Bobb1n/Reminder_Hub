package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	postgresConfig "auth/pkg/postgres"
)

type Config struct {
	Port      int    `env:"SERVER_PORT" env-default:"8081"`
	JWTSecret string `env:"JWT_SECRET"`

	postgresConfig.Config
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
