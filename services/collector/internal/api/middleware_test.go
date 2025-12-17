package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestUserIDAuth_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := UserIDAuth(func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	if err := h(c); err == nil {
		t.Fatal("expected error when X-User-ID header is missing")
	}
}

func TestUserIDAuth_InvalidUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "not-a-uuid")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := UserIDAuth(func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	if err := h(c); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestUserIDAuth_SetsContext(t *testing.T) {
	e := echo.New()
	validID := "123e4567-e89b-12d3-a456-426614174000"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", validID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	h := UserIDAuth(func(c echo.Context) error {
		called = true
		if got := c.Get(ContextKeyUserID); got != validID {
			t.Fatalf("Context user_id = %v, want %v", got, validID)
		}
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}
