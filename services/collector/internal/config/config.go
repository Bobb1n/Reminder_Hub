package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	DBURL            string
	RabbitURL        string
	QueueName        string
	ServerPort       string
	ServerHost       string
	InternalAPIToken string
}

func Load() (*Config, error) {
	if _, err := os.Stat("../../../../.env"); err == nil {
		_ = godotenv.Load("../../../../.env")
	} else {
		log.Error().Err(err)
	}

	cfg := &Config{
		DBURL:            env("DB_URL", ""),
		RabbitURL:        env("RABBIT_URL", ""),
		QueueName:        env("RABBIT_QUEUE_NAME", "parsed_emails_queue"),
		ServerPort:       env("SERVER_PORT", "8080"),
		ServerHost:       env("SERVER_HOST", "0.0.0.0"),
		InternalAPIToken: env("INTERNAL_API_TOKEN", "gateway-secret-token"),
	}

	if cfg.DBURL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	if cfg.RabbitURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL is required")
	}

	log.Info().
		Str("ServerPort", cfg.ServerPort).
		Str("RabbitURL", cfg.RabbitURL).
		Str("QueueName", cfg.QueueName).
		Str("DBURL", cfg.DBURL).
		Msg("Config loaded")

	return cfg, nil
}

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
