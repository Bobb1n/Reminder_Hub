package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"reminder-hub/pkg/logger"

	"github.com/labstack/echo/v4"
)

type nopLogger struct{}

func (nopLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (nopLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (nopLogger) Panic(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func newTestLogger() *logger.CurrentLogger { return logger.NewCurrentLogger(nopLogger{}) }

func TestNewServiceProxy_ServiceTypeDetection(t *testing.T) {
	log := newTestLogger()
	p, err := NewServiceProxy("http://core-service", "token", log)
	if err != nil || p.serviceType != "core" {
		t.Fatalf("core serviceType=%q err=%v", p.serviceType, err)
	}

	p, err = NewServiceProxy("http://collector-service", "token", log)
	if err != nil || p.serviceType != "collector" {
		t.Fatalf("collector serviceType=%q err=%v", p.serviceType, err)
	}
}

func TestAuthProxy_InvalidURL(t *testing.T) {
	if _, err := AuthProxy("://bad"); err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestAuthProxy_Serves(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer backend.Close()

	h, err := AuthProxy(backend.URL)
	if err != nil {
		t.Fatalf("AuthProxy error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestServiceProxy_ProxySetsHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got == "" {
			t.Errorf("missing Authorization header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	log := newTestLogger()
	p, err := NewServiceProxy(backend.URL, "internal", log)
	if err != nil {
		t.Fatalf("NewServiceProxy error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/email", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "u1")

	if err := p.Proxy(c); err != nil {
		t.Fatalf("Proxy error: %v", err)
	}
}
