package config

import (
	"os"
	"testing"
)

func TestLoad_RequiresJWTSecret(t *testing.T) {
	os.Clearenv()
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}

	os.Setenv("JWT_SECRET", "secret")
	defer os.Unsetenv("JWT_SECRET")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.JWTSecret != "secret" || cfg.Logger == nil {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
