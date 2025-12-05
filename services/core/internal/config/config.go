package config

import (
	"os"
	"path/filepath"
	"strconv"
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
	loadEnv()

	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8082"),
		DBURL:            getEnv("CORE_DB_URL", "postgres://reminder:reminder@localhost:5432/reminderhub?sslmode=disable"),
		RabbitURL:        getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		SyncInterval:     getDur("SYNC_INTERVAL", 30*time.Second),
		IMAPTimeout:      getDur("IMAP_TIMEOUT", 30*time.Second),
		MaxWorkers:       getInt("MAX_WORKERS", 10),
		BatchSize:        getInt("BATCH_SIZE", 50),
		EncryptionKey:    getEnv("ENCRYPTION_KEY", "default-32-char-encryption-key-here!!"),
		InternalAPIToken: getEnv("INTERNAL_API_TOKEN", "gateway-secret-token"),
	}

	if len(cfg.EncryptionKey) != 32 {
		if len(cfg.EncryptionKey) < 32 {
			cfg.EncryptionKey += string(make([]byte, 32-len(cfg.EncryptionKey)))
		} else {
			cfg.EncryptionKey = cfg.EncryptionKey[:32]
		}
		log.Warn().Msg("ENCRYPTION_KEY adjusted to 32 chars")
	}

	return cfg
}

func loadEnv() {
	paths := []string{".env", "../.env", "../../.env", "../../../.env"}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		paths = append(paths, filepath.Join(dir, ".env"))
		paths = append(paths, filepath.Join(dir, "..", ".env"))
	}

	if custom := os.Getenv("ENV_FILE_PATH"); custom != "" {
		paths = append([]string{custom}, paths...)
	}

	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			log.Info().Msgf("Loaded: %s", p)
			return
		}
	}
	log.Info().Msg("No .env found, using env vars")
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return d
}

func getDur(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			return dur
		}
	}
	return d
}
