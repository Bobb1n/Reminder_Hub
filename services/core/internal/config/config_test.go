package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"reminder-hub/pkg/rabbitmq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

type simpleLifecycle struct{}

func (s *simpleLifecycle) Append(hook fx.Hook) {}

func TestGet_WithEnvVar(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	result := get("TEST_KEY", "default")
	assert.Equal(t, "test_value", result)
}

func TestGet_WithDefault(t *testing.T) {
	result := get("NON_EXISTENT_KEY", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetInt_WithEnvVar(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getInt("TEST_INT", 10)
	assert.Equal(t, 42, result)
}

func TestGetInt_WithInvalidEnvVar(t *testing.T) {
	os.Setenv("TEST_INT", "invalid")
	defer os.Unsetenv("TEST_INT")

	result := getInt("TEST_INT", 10)
	assert.Equal(t, 10, result) // Должен вернуть default
}

func TestGetInt_WithDefault(t *testing.T) {
	result := getInt("NON_EXISTENT_INT", 99)
	assert.Equal(t, 99, result)
}

func TestGetDuration_WithEnvVar(t *testing.T) {
	os.Setenv("TEST_DURATION", "5s")
	defer os.Unsetenv("TEST_DURATION")

	result := getDuration("TEST_DURATION", 10*time.Second)
	assert.Equal(t, 5*time.Second, result)
}

func TestGetDuration_WithInvalidEnvVar(t *testing.T) {
	os.Setenv("TEST_DURATION", "invalid")
	defer os.Unsetenv("TEST_DURATION")

	result := getDuration("TEST_DURATION", 10*time.Second)
	assert.Equal(t, 10*time.Second, result) // Должен вернуть default
}

func TestGetDuration_WithDefault(t *testing.T) {
	result := getDuration("NON_EXISTENT_DURATION", 30*time.Second)
	assert.Equal(t, 30*time.Second, result)
}

func TestNormalizeKey_WithLongKey(t *testing.T) {
	longKey := "this_is_a_very_long_key_that_exceeds_32_characters"
	result := normalizeKey(longKey)
	assert.Equal(t, 32, len(result))
	assert.Equal(t, longKey[:32], result)
}

func TestNormalizeKey_WithShortKey(t *testing.T) {
	shortKey := "short"
	result := normalizeKey(shortKey)
	assert.Equal(t, 32, len(result))
	assert.True(t, strings.HasPrefix(result, shortKey))
}

func TestNormalizeKey_WithExactLength(t *testing.T) {
	exactKey := "12345678901234567890123456789012" // Ровно 32 символа
	result := normalizeKey(exactKey)
	assert.Equal(t, 32, len(result))
	assert.Equal(t, exactKey, result)
}

func TestMask_WithValidURL(t *testing.T) {
	url := "postgres://user:password@localhost:5432/dbname"
	result := mask(url)
	assert.Contains(t, result, "user")
	assert.Contains(t, result, "***")
	assert.Contains(t, result, "localhost:5432/dbname")
	assert.NotContains(t, result, "password")
}

func TestMask_WithInvalidURL(t *testing.T) {
	url := "invalid_url"
	result := mask(url)
	assert.Equal(t, url, result) // Должен вернуть исходный URL
}

func TestParseRabbitURL_WithFullURL(t *testing.T) {
	cfg := &rabbitmq.RabbitMQConfig{}
	url := "amqp://user:pass@host:5672/vhost"
	parseRabbitURL(url, cfg)

	assert.Equal(t, "user", cfg.User)
	assert.Equal(t, "pass", cfg.Password)
	assert.Equal(t, "host", cfg.Host)
	assert.Equal(t, 5672, cfg.Port)
}

func TestParseRabbitURL_WithNoPort(t *testing.T) {
	cfg := &rabbitmq.RabbitMQConfig{}
	url := "amqp://user:pass@host/vhost"
	parseRabbitURL(url, cfg)

	assert.Equal(t, "user", cfg.User)
	assert.Equal(t, "pass", cfg.Password)
	assert.Equal(t, "host", cfg.Host)
}

func TestParseRabbitURL_WithInvalidURL(t *testing.T) {
	cfg := &rabbitmq.RabbitMQConfig{
		User: "default_user",
		Host: "default_host",
	}
	url := "invalid_url"
	parseRabbitURL(url, cfg)

	// Должны остаться значения по умолчанию
	assert.Equal(t, "default_user", cfg.User)
	assert.Equal(t, "default_host", cfg.Host)
}

func TestLoadRabbitMQConfig_WithDefaults(t *testing.T) {
	// Очищаем переменные окружения
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_USER")
	os.Unsetenv("RABBITMQ_PASSWORD")
	os.Unsetenv("RABBIT_URL")

	cfg := loadRabbitMQConfig()

	assert.Equal(t, "rabbitmq", cfg.Host)
	assert.Equal(t, 5672, cfg.Port)
	assert.Equal(t, "guest", cfg.User)
	assert.Equal(t, "guest", cfg.Password)
}

func TestLoadRabbitMQConfig_WithEnvVars(t *testing.T) {
	os.Setenv("RABBITMQ_HOST", "custom_host")
	os.Setenv("RABBITMQ_PORT", "1234")
	os.Setenv("RABBITMQ_USER", "custom_user")
	os.Setenv("RABBITMQ_PASSWORD", "custom_pass")
	defer func() {
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USER")
		os.Unsetenv("RABBITMQ_PASSWORD")
	}()

	cfg := loadRabbitMQConfig()

	assert.Equal(t, "custom_host", cfg.Host)
	assert.Equal(t, 1234, cfg.Port)
	assert.Equal(t, "custom_user", cfg.User)
	assert.Equal(t, "custom_pass", cfg.Password)
}

func TestLoad_WithDefaults(t *testing.T) {
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)

	cfg := Load(logger)

	require.NotNil(t, cfg)
	assert.Equal(t, "8082", cfg.ServerPort)
	assert.Equal(t, 30*time.Second, cfg.SyncInterval)
	assert.Equal(t, 30*time.Second, cfg.IMAPTimeout)
	assert.Equal(t, 10, cfg.MaxWorkers)
	assert.Equal(t, 50, cfg.BatchSize)
}
