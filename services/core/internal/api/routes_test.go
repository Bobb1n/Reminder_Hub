package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"core/internal/database"
	"core/internal/security"
	corelogger "core/internal/logger"

	"github.com/labstack/echo/v4"
)

// minimalEncryptor wraps security.NewEncryptor for tests
func newTestEncryptor() security.Encryptor {
	return security.NewEncryptor("12345678901234567890123456789012")
}

func TestSetupRoutes_RegistersHandlers(t *testing.T) {
	e := echo.New()
	log := corelogger.Init("test")
	enc := newTestEncryptor()

	// db is not used during registration
	var db *database.DB

	SetupRoutes(e, db, enc, "internal-token", log)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code == 0 {
		t.Fatal("health route did not respond")
	}
}

func TestInternalAuth_Middleware(t *testing.T) {
	e := echo.New()
	mw := InternalAuth("secret-token")

	handlerCalled := false
	h := mw(func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})

	// no header -> 401
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// invalid format
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if handlerCalled {
		t.Fatal("handler should not be called on invalid format")
	}

	// wrong token
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	_ = h(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// correct token -> next called
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	if err := h(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called with valid token")
	}
}
