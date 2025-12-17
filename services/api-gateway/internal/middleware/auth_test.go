package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newTestLogger() *logger.CurrentLogger {
	return logger.NewCurrentLogger(nopLogger{})
}

func TestAuthMiddleware_SkipsHealthAndAuthPaths(t *testing.T) {
	e := echo.New()
	log := newTestLogger()
	mw := AuthMiddleware("http://auth", log)

	called := false
	next := func(c echo.Context) error { called = true; return c.NoContent(http.StatusOK) }
	h := mw(next)

	for _, path := range []string{"/health", "/auth/login"} {
		called = false
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)

		if err := h(c); err != nil {
			t.Fatalf("unexpected error for %s: %v", path, err)
		}
		if !called {
			t.Fatalf("next not called for path %s", path)
		}
	}
}

func TestParseValidateResponse(t *testing.T) {
	body := `{"valid":true,"user_id":"u1","email":"a@b"}`
	id, err := parseValidateResponse(bytes.NewBufferString(body))
	if err != nil || id != "u1" {
		t.Fatalf("parseValidateResponse = %q, %v", id, err)
	}

	body = `{"valid":false}`
	if _, err := parseValidateResponse(strings.NewReader(body)); err == nil {
		t.Fatal("expected error for invalid token")
	}

	bad := io.NopCloser(strings.NewReader("{"))
	defer bad.Close()
	if _, err := parseValidateResponse(bad); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestHandleErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusUnauthorized)
	_, _ = rec.WriteString("unauthorized")

	err := handleErrorResponse(rec.Result())
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToken_ErrorPropagates(t *testing.T) {
	log := newTestLogger()
	ctx := context.Background()

	_, err := validateToken("http://bad-host", "token", log, ctx)
	if err == nil {
		t.Fatal("expected error from validateToken for bad host")
	}
}
