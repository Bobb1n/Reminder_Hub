package config

import (
	"os"
	"testing"
)

func TestGetEnv_DefaultAndOverride(t *testing.T) {
	const key = "TEST_API_GW_ENV_KEY"
	os.Unsetenv(key)
	if v := getEnv(key, "default"); v != "default" {
		t.Fatalf("getEnv default = %q, want %q", v, "default")
	}

	os.Setenv(key, "custom")
	defer os.Unsetenv(key)
	if v := getEnv(key, "default"); v != "custom" {
		t.Fatalf("getEnv override = %q, want %q", v, "custom")
	}
}

func TestLoad_UsesDefaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.ServerPort == "" || cfg.Logger == nil {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
