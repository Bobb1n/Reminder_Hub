package config

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/rabbitmq"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment      string
	ServerPort       string
	DBURL            string
	RabbitURL        string
	Rabbitmq         *rabbitmq.RabbitMQConfig
	SyncInterval     time.Duration
	IMAPTimeout      time.Duration
	MaxWorkers       int
	BatchSize        int
	EncryptionKey    string
	InternalAPIToken string
}

func Load(appLogger *logger.CurrentLogger) *Config {
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load(".env")
	}

	cfg := &Config{
		Environment:      get("ENV", "development"),
		ServerPort:       get("SERVER_PORT", "8082"),
		DBURL:            get("CORE_DB_URL", "postgres://reminder:reminder@postgres:5432/reminderhub?sslmode=disable"),
		RabbitURL:        get("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/"),
		SyncInterval:     getDuration("SYNC_INTERVAL", 30*time.Second),
		IMAPTimeout:      getDuration("IMAP_TIMEOUT", 30*time.Second),
		MaxWorkers:       getInt("MAX_WORKERS", 10),
		BatchSize:        getInt("BATCH_SIZE", 50),
		EncryptionKey:    normalizeKey(get("ENCRYPTION_KEY", "fV6dIefy6ViClzMX0wYC+fXJf3smOuAI")),
		InternalAPIToken: get("INTERNAL_API_TOKEN", "gateway-secret-token"),
		Rabbitmq:         loadRabbitMQConfig(),
	}

	if appLogger != nil {
		logConfig(appLogger, cfg)
	}

	return cfg
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func normalizeKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	return key + strings.Repeat("0", 32-len(key))
}

func mask(url string) string {
	parts := strings.SplitN(url, "@", 2)
	if len(parts) != 2 {
		return url
	}

	auth := strings.SplitN(parts[0], "://", 2)
	if len(auth) != 2 {
		return url
	}

	cred := strings.SplitN(auth[1], ":", 2)
	if len(cred) != 2 {
		return url
	}

	return auth[0] + "://" + cred[0] + ":***@" + parts[1]
}

func loadRabbitMQConfig() *rabbitmq.RabbitMQConfig {
	cfg := &rabbitmq.RabbitMQConfig{
		Host:         get("RABBITMQ_HOST", "rabbitmq"),
		Port:         getInt("RABBITMQ_PORT", 5672),
		User:         get("RABBITMQ_USER", "guest"),
		Password:     get("RABBITMQ_PASSWORD", "guest"),
		ExchangeName: get("RABBITMQ_EXCHANGE", "donotmatter"),
		Kind:         get("RABBITMQ_KIND", "topic"),
	}

	if rabbitURL := get("RABBIT_URL", ""); rabbitURL != "" && strings.HasPrefix(rabbitURL, "amqp://") {
		parseRabbitURL(rabbitURL, cfg)
	}

	return cfg
}

func parseRabbitURL(url string, cfg *rabbitmq.RabbitMQConfig) {
	urlWithoutScheme := strings.TrimPrefix(url, "amqp://")
	parts := strings.SplitN(urlWithoutScheme, "@", 2)
	if len(parts) != 2 {
		return
	}

	auth := strings.SplitN(parts[0], ":", 2)
	if len(auth) == 2 {
		cfg.User = auth[0]
		cfg.Password = auth[1]
	}

	hostPort := strings.SplitN(strings.Split(parts[1], "/")[0], ":", 2)
	if len(hostPort) > 0 {
		cfg.Host = hostPort[0]
	}
	if len(hostPort) == 2 {
		if port, err := strconv.Atoi(hostPort[1]); err == nil {
			cfg.Port = port
		}
	}
}

func logConfig(appLogger *logger.CurrentLogger, cfg *Config) {
	ctx := context.Background()
	appLogger.Info(ctx, "Config loaded",
		"ServerPort", cfg.ServerPort,
		"RabbitURL", mask(cfg.RabbitURL),
		"DBURL", mask(cfg.DBURL),
		"SyncInterval", cfg.SyncInterval.String(),
		"IMAPTimeout", cfg.IMAPTimeout.String(),
		"MaxWorkers", cfg.MaxWorkers,
		"BatchSize", cfg.BatchSize,
	)
}
