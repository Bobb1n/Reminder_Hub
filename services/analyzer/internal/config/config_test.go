package config

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	os.Unsetenv("RABBITMQ_HOST")
	if v := getEnvOrDefault("RABBITMQ_HOST", "localhost"); v != "localhost" {
		t.Fatalf("default host = %q", v)
	}
	os.Setenv("RABBITMQ_HOST", "example")
	defer os.Unsetenv("RABBITMQ_HOST")
	if v := getEnvOrDefault("RABBITMQ_HOST", "localhost"); v != "example" {
		t.Fatalf("env host = %q", v)
	}
}

func TestParseRabbitMQConfig_DefaultPort(t *testing.T) {
	os.Unsetenv("RABBITMQ_PORT")
	cfg := parseRabbitMQConfig()
	if cfg.Port != 5672 {
		t.Fatalf("default port = %d", cfg.Port)
	}
}

func TestGetMicroserviceName(t *testing.T) {
	if got := GetMicroserviceName("analyzer"); got != "ANALYZER" {
		t.Fatalf("GetMicroserviceName = %q", got)
	}
}
