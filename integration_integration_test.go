//go:build integration

package integration

import (
	"net/http"
	"os"
	"testing"
	"time"
)

// helper to get env with default
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// TestSystemHealth checks that all HTTP services respond on their health endpoints.
// It assumes the system is already running via docker-compose or manually.
func TestSystemHealth(t *testing.T) {
	t.Parallel()

	client := &http.Client{Timeout: 5 * time.Second}

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "api-gateway health",
			url:  env("API_GATEWAY_URL", "http://localhost:8080") + "/health",
		},
		{
			name: "auth service health",
			url:  env("AUTH_SERVICE_URL", "http://localhost:8081") + "/health",
		},
		{
			name: "core service health",
			url:  env("CORE_SERVICE_URL", "http://localhost:8082") + "/api/health",
		},
		{
			name: "collector service health",
			url:  env("COLLECTOR_SERVICE_URL", "http://localhost:8083") + "/health",
		},
		{
			name: "analyzer service health",
			url:  env("ANALYZER_SERVICE_URL", "http://localhost:8084") + "/health",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(tt.url)
			if err != nil {
				t.Fatalf("GET %s error: %v", tt.url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				t.Fatalf("%s returned status %d", tt.url, resp.StatusCode)
			}
		})
	}
}
