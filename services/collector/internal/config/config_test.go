package config

import (
	"os"
	"testing"
)

func TestEnv_DefaultAndOverride(t *testing.T) {
	const key = "COLLECTOR_TEST_KEY"
	os.Unsetenv(key)
	if v := env(key, "def"); v != "def" {
		t.Fatalf("env default = %q, want %q", v, "def")
	}
	os.Setenv(key, "custom")
	defer os.Unsetenv(key)
	if v := env(key, "def"); v != "custom" {
		t.Fatalf("env override = %q, want %q", v, "custom")
	}
}

func TestLoad_RequiresDBAndRabbitURL(t *testing.T) {
	// We only assert that when required envs are present, Load succeeds.
	os.Setenv("DB_URL", "postgres://user:pass@localhost/db")
	os.Setenv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	defer os.Unsetenv("DB_URL")
	defer os.Unsetenv("RABBIT_URL")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.DBURL == "" || cfg.RabbitURL == "" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
