package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	ServerPort       string
	DBURL            string
	RabbitURL        string
	SyncInterval     time.Duration
	IMAPTimeout      time.Duration
	MaxWorkers       int
	BatchSize        int
	EncryptionKey    string
	InternalAPIToken string
}

func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load(".env")
	}

	cfg := &Config{
		ServerPort:       env("SERVER_PORT", "8082"),
		DBURL:            env("CORE_DB_URL", "postgres://reminder:reminder@postgres:5432/reminderhub?sslmode=disable"),
		RabbitURL:        env("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/"),
		SyncInterval:     dur("SYNC_INTERVAL", 30*time.Second),
		IMAPTimeout:      dur("IMAP_TIMEOUT", 30*time.Second),
		MaxWorkers:       num("MAX_WORKERS", 10),
		BatchSize:        num("BATCH_SIZE", 50),
		EncryptionKey:    env("ENCRYPTION_KEY", "fV6dIefy6ViClzMX0wYC+fXJf3smOuAI"),
		InternalAPIToken: env("INTERNAL_API_TOKEN", "gateway-secret-token"),
	}

	if len(cfg.EncryptionKey) != 32 {
		if len(cfg.EncryptionKey) < 32 {
			cfg.EncryptionKey += string(make([]byte, 32-len(cfg.EncryptionKey)))
		} else {
			cfg.EncryptionKey = cfg.EncryptionKey[:32]
		}
	}

	log.Info().
		Str("ServerPort", cfg.ServerPort).
		Str("RabbitURL", mask(cfg.RabbitURL)).
		Str("DBURL", mask(cfg.DBURL)).
		Dur("SyncInterval", cfg.SyncInterval).
		Dur("IMAPTimeout", cfg.IMAPTimeout).
		Int("MaxWorkers", cfg.MaxWorkers).
		Int("BatchSize", cfg.BatchSize).
		Msg("Config loaded")

	return cfg
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func num(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func dur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func mask(url string) string {
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return url
	}
	auth := strings.Split(parts[0], "://")
	if len(auth) != 2 {
		return url
	}
	cred := strings.Split(auth[1], ":")
	if len(cred) != 2 {
		return url
	}
	return auth[0] + "://" + cred[0] + ":***@" + parts[1]
}
