package config

import (
	"os"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"

	"go.uber.org/fx"
)

type Config struct {
	ServerPort          string
	AuthServiceURL      string
	CoreServiceURL      string
	CollectorServiceURL string
	InternalToken       string
	JWTSecret           string
	Logger              *logger.CurrentLogger
}

func Load() (*Config, error) {
	env := getEnv("ENV", "development")

	// Создаем минимальный lifecycle для логгера
	lc := &simpleLifecycle{}

	adapter := zaplogger.NewLoggerAdapter(lc, env)

	cfg := &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		AuthServiceURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		CoreServiceURL:      getEnv("CORE_SERVICE_URL", "http://core-service:8082"),
		CollectorServiceURL: getEnv("COLLECTOR_SERVICE_URL", "http://collector-service:8083"),
		InternalToken:       getEnv("INTERNAL_API_TOKEN", "gateway-secret-token"),
		JWTSecret:           getEnv("JWT_SECRET", "your-jwt-secret-key"),
		Logger:              logger.NewCurrentLogger(adapter),
	}

	return cfg, nil
}

// simpleLifecycle - упрощенная реализация fx.Lifecycle для сервисов без Fx
type simpleLifecycle struct {
	hooks []fx.Hook
}

func (lc *simpleLifecycle) Append(hook fx.Hook) {
	lc.hooks = append(lc.hooks, hook)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
