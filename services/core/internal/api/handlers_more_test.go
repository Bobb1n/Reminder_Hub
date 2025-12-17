package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"core/internal/database"
	corelogger "core/internal/logger"

	"github.com/labstack/echo/v4"
)

func TestHandler_GetUserIntegrations_DBError(t *testing.T) {
	log := corelogger.Init("test")
	mdb := &mockDB{}
	mdb.getUserIntegrationsFunc = func(ctx context.Context, userID string) ([]database.EmailIntegration, error) {
		return nil, errors.New("database error")
	}
	h := NewHandler(mdb, &mockEncryptor{}, log)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/integrations/user-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("user_id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	if err := h.GetUserIntegrations(c); err != nil {
		t.Fatalf("GetUserIntegrations error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandler_DeleteIntegration_NotFound(t *testing.T) {
	log := corelogger.Init("test")
	mdb := &mockDB{}
	mdb.deleteIntegrationFunc = func(ctx context.Context, userID, integrationID string) error {
		return database.ErrIntegrationNotFound
	}
	h := NewHandler(mdb, &mockEncryptor{}, log)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/integrations/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
	c.Set(ContextKeyUserID, "user-id")

	if err := h.DeleteIntegration(c); err != nil {
		t.Fatalf("DeleteIntegration error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandler_DeleteIntegration_DBError(t *testing.T) {
	log := corelogger.Init("test")
	mdb := &mockDB{}
	mdb.deleteIntegrationFunc = func(ctx context.Context, userID, integrationID string) error {
		return errors.New("database error")
	}
	h := NewHandler(mdb, &mockEncryptor{}, log)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/integrations/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
	c.Set(ContextKeyUserID, "user-id")

	if err := h.DeleteIntegration(c); err != nil {
		t.Fatalf("DeleteIntegration error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}


