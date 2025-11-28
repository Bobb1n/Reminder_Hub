package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBURL       string
	RabbitMQURL string
	ServerPort  string
}

func Load() *Config {
	return &Config{
		DBURL:       getEnv("DB_URL", "postgres://reminder:reminder@postgres:5432/reminderhub?sslmode=disable"),
		RabbitMQURL: getEnv("RABBIT_URL", "amqp://rabbit:rabbit@rabbitmq:5672/"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
