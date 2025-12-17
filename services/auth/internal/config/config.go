package config

import (
	"fmt"
	"os"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/fx"

	postgresConfig "auth/pkg/postgres"
)

type Config struct {
	Port      int    `env:"SERVER_PORT" env-default:"8081"`
	JWTSecret string `env:"JWT_SECRET"`

	RedisHost     string `env:"REDIS_HOST" env-default:"localhost"`
	RedisPort     int    `env:"REDIS_PORT" env-default:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" env-default:""`
	RedisDB       int    `env:"REDIS_DB" env-default:"0"`

	postgresConfig.Config
	Logger *logger.CurrentLogger
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, env)
	cfg.Logger = logger.NewCurrentLogger(adapter)

	return cfg, nil
}

type simpleLifecycle struct {
	hooks []fx.Hook
}

func (lc *simpleLifecycle) Append(hook fx.Hook) {
	lc.hooks = append(lc.hooks, hook)
}
