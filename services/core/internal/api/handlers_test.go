package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"core/internal/database"
	"core/internal/util"
	corelogger "core/internal/logger"

	"github.com/labstack/echo/v4"
)

type mockDB struct{
	createIntegrationFunc func(ctx context.Context, integration *database.EmailIntegration) error
}

func (m *mockDB) CreateIntegration(ctx context.Context, integration *database.EmailIntegration) error {
	if m.createIntegrationFunc != nil {
		return m.createIntegrationFunc(ctx, integration)
	}
	return nil
}

func (m *mockDB) GetUserIntegrations(ctx context.Context, userID string) ([]database.EmailIntegration, error) {
	return nil, nil
}
func (m *mockDB) DeleteIntegration(ctx context.Context, userID, integrationID string) error {
	return nil
}
func (m *mockDB) GetIntegrationsForSync(ctx context.Context, limit int) ([]database.EmailIntegration, error) {
	return nil, nil
}
func (m *mockDB) UpdateLastSync(ctx context.Context, integrationID string) error {
	return nil
}
func (m *mockDB) EmailExists(ctx context.Context, userID, messageID string) (bool, error) {
	return false, nil
}
func (m *mockDB) SaveEmail(ctx context.Context, email *database.EmailRaw) error {
	return nil
}

type mockEncryptor struct{
	encryptFunc func(text string) (string, error)
}

func (m *mockEncryptor) Encrypt(text string) (string, error) {
	if m.encryptFunc != nil {
		return m.encryptFunc(text)
	}
	return text, nil
}

func (m *mockEncryptor) Decrypt(cipherText string) (string, error) {
	return cipherText, nil
}

func TestHandler_HealthCheck(t *testing.T) {
	e := echo.New()
	log := corelogger.Init("test")
	h := NewHandler(&mockDB{}, &mockEncryptor{}, log)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HealthCheck(c); err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_encryptPassword(t *testing.T) {
	log := corelogger.Init("test")
	me := &mockEncryptor{}
	h := NewHandler(&mockDB{}, me, log)
	ctx := context.Background()

	me.encryptFunc = func(text string) (string, error) { return "enc-" + text, nil }
	res, err := h.encryptPassword(ctx, "pwd")
	if err != nil || res != "enc-pwd" {
		t.Fatalf("encryptPassword result=%q err=%v", res, err)
	}

	me.encryptFunc = func(text string) (string, error) { return "", errors.New("fail") }
	if _, err := h.encryptPassword(ctx, "pwd"); err == nil {
		t.Fatal("expected error on encryption failure")
	}
}

func TestCreateIntegrationRecord_Normalizes(t *testing.T) {
	log := corelogger.Init("test")
	h := NewHandler(&mockDB{}, &mockEncryptor{}, log)
	ctx := context.Background()

	req := &database.CreateIntegrationRequest{EmailAddress: "  TeSt@ExAmPlE.Com  ", ImapHost: "imap", ImapPort: 993, UseSSL: true}
	id, _ := util.GenerateUUID()
	_ = id

	integration, err := h.createIntegrationRecord(ctx, req, "user-id", "enc")
	if err != nil {
		t.Fatalf("createIntegrationRecord error: %v", err)
	}
	if integration.EmailAddress != "test@example.com" {
		t.Fatalf("normalized email = %q, want %q", integration.EmailAddress, "test@example.com")
	}
}

func TestBindAndValidate_InvalidBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var r database.CreateIntegrationRequest
	if err := bindAndValidate(c, &r); err == nil {
		t.Fatal("expected error on invalid body (empty)")
	}
}
