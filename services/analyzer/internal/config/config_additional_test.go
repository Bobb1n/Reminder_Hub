package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRabbitMQConfig_WithEnvVars(t *testing.T) {
	os.Setenv("RABBITMQ_HOST", "custom_host")
	os.Setenv("RABBITMQ_PORT", "1234")
	os.Setenv("RABBITMQ_USER", "custom_user")
	os.Setenv("RABBITMQ_PASSWORD", "custom_pass")
	os.Setenv("RABBITMQ_EXCHANGE", "custom_exchange")
	os.Setenv("RABBITMQ_KIND", "direct")
	defer func() {
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USER")
		os.Unsetenv("RABBITMQ_PASSWORD")
		os.Unsetenv("RABBITMQ_EXCHANGE")
		os.Unsetenv("RABBITMQ_KIND")
	}()

	cfg := parseRabbitMQConfig()

	assert.Equal(t, "custom_host", cfg.Host)
	assert.Equal(t, 1234, cfg.Port)
	assert.Equal(t, "custom_user", cfg.User)
	assert.Equal(t, "custom_pass", cfg.Password)
	assert.Equal(t, "custom_exchange", cfg.ExchangeName)
	assert.Equal(t, "direct", cfg.Kind)
}

func TestParseRabbitMQConfig_InvalidPort(t *testing.T) {
	os.Setenv("RABBITMQ_PORT", "invalid")
	defer os.Unsetenv("RABBITMQ_PORT")

	cfg := parseRabbitMQConfig()

	// Должен использовать значение по умолчанию
	assert.Equal(t, 5672, cfg.Port)
}

func TestParseRabbitMQConfig_WithDefaults(t *testing.T) {
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_USER")
	os.Unsetenv("RABBITMQ_PASSWORD")
	os.Unsetenv("RABBITMQ_EXCHANGE")
	os.Unsetenv("RABBITMQ_KIND")

	cfg := parseRabbitMQConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5672, cfg.Port)
	assert.Equal(t, "guest", cfg.User)
	assert.Equal(t, "guest", cfg.Password)
	assert.Equal(t, "donotmatter", cfg.ExchangeName)
	assert.Equal(t, "topic", cfg.Kind)
}

func TestGetEnvOrDefault_WithValue(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	result := getEnvOrDefault("TEST_KEY", "default")
	assert.Equal(t, "test_value", result)
}

func TestGetEnvOrDefault_WithDefault(t *testing.T) {
	os.Unsetenv("TEST_KEY")

	result := getEnvOrDefault("TEST_KEY", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetMicroserviceName_DifferentCases(t *testing.T) {
	assert.Equal(t, "ANALYZER", GetMicroserviceName("analyzer"))
	assert.Equal(t, "CORE", GetMicroserviceName("core"))
	assert.Equal(t, "AUTH", GetMicroserviceName("auth"))
	assert.Equal(t, "COLLECTOR", GetMicroserviceName("collector"))
}

func TestGetMicroserviceName_WithMixedCase(t *testing.T) {
	assert.Equal(t, "ANALYZER", GetMicroserviceName("AnAlYzEr"))
	assert.Equal(t, "TEST", GetMicroserviceName("TeSt"))
}

