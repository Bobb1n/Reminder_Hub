package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"collector/internal/service"

	"github.com/labstack/echo/v4"
)

func TestSetupRoutes_HealthRegistered(t *testing.T) {
	e := echo.New()
	ts := &service.TaskService{}
	SetupRoutes(e, ts, "token")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == 0 {
		t.Fatal("health route did not respond")
	}
}

func TestInternalAuth_Middleware(t *testing.T) {
	e := echo.New()
	mw := InternalAuth("secret")

	h := mw(func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if err := h(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
}
