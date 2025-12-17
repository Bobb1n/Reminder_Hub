package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"reminder-hub/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockLogger - простой мок-логгер для тестов
type MockLogger struct{}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (m *MockLogger) Panic(ctx context.Context, msg string, fields ...any) {}
func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func setupTestLogger() *logger.CurrentLogger {
	return logger.NewCurrentLogger(&MockLogger{})
}

func setupTestBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки
		authHeader := r.Header.Get("Authorization")
		internalToken := r.Header.Get("X-Internal-Token")
		userID := r.Header.Get("X-User-ID")

		response := map[string]interface{}{
			"path":           r.URL.Path,
			"method":         r.Method,
			"auth_header":    authHeader,
			"internal_token": internalToken,
			"user_id":        userID,
		}

		// Если есть body, читаем его
		if r.Body != nil {
			var bodyData map[string]interface{}
			json.NewDecoder(r.Body).Decode(&bodyData)
			response["body"] = bodyData
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func TestServiceProxy_Proxy_IntegrationPathRewrite(t *testing.T) {
	logger := setupTestLogger()
	backend := setupTestBackend()
	defer backend.Close()

	proxy, err := NewServiceProxy(backend.URL, "internal-token-123", logger)
	assert.NoError(t, err)

	body := map[string]string{
		"email_address": "test@example.com",
		"password":      "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/email", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/integrations/email")
	c.Set("user_id", "test-user-123")

	err = proxy.Proxy(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "/api/integrations", response["path"])
	assert.Equal(t, "Bearer internal-token-123", response["auth_header"])
	assert.Equal(t, "internal-token-123", response["internal_token"])
	assert.Equal(t, "test-user-123", response["user_id"])
}

func TestServiceProxy_Proxy_IntegrationGETPathRewrite(t *testing.T) {
	logger := setupTestLogger()
	backend := setupTestBackend()
	defer backend.Close()

	proxy, err := NewServiceProxy(backend.URL, "internal-token-123", logger)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/email", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/integrations/email")
	c.Set("user_id", "test-user-123")

	err = proxy.Proxy(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "/api/integrations/test-user-123", response["path"])
}

func TestServiceProxy_Proxy_RemindersPathRewrite(t *testing.T) {
	logger := setupTestLogger()
	backend := setupTestBackend()
	defer backend.Close()

	// Создаем proxy для collector-service (URL содержит "collector")
	proxy, err := NewServiceProxy(backend.URL, "internal-token-123", logger)
	assert.NoError(t, err)
	// Принудительно устанавливаем serviceType для теста
	proxy.serviceType = "collector"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reminders/tasks", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/reminders/tasks")
	c.Set("user_id", "test-user-123")

	err = proxy.Proxy(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	// Путь должен быть переписан с /api/v1/reminders/tasks на /api/v1/tasks
	assert.Equal(t, "/api/v1/tasks", response["path"])
}

func TestServiceProxy_Proxy_HeadersSet(t *testing.T) {
	logger := setupTestLogger()
	backend := setupTestBackend()
	defer backend.Close()

	proxy, err := NewServiceProxy(backend.URL, "internal-token-123", logger)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/test")
	c.Set("user_id", "test-user-123")

	err = proxy.Proxy(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Bearer internal-token-123", response["auth_header"])
	assert.Equal(t, "internal-token-123", response["internal_token"])
	assert.Equal(t, "test-user-123", response["user_id"])
}

func TestServiceProxy_Proxy_BodyCopied(t *testing.T) {
	logger := setupTestLogger()
	backend := setupTestBackend()
	defer backend.Close()

	proxy, err := NewServiceProxy(backend.URL, "internal-token-123", logger)
	assert.NoError(t, err)

	body := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/test")

	err = proxy.Proxy(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Проверяем, что body был передан
	bodyData, ok := response["body"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", bodyData["field1"])
	assert.Equal(t, "value2", bodyData["field2"])
}

func TestServiceProxy_NewServiceProxy_InvalidURL(t *testing.T) {
	logger := setupTestLogger()
	_, err := NewServiceProxy("://invalid-url", "token", logger)
	assert.Error(t, err)
}

func TestServiceProxy_NewServiceProxy_CollectorServiceType(t *testing.T) {
	logger := setupTestLogger()
	proxy, err := NewServiceProxy("http://collector-service:8083", "token", logger)
	assert.NoError(t, err)
	assert.Equal(t, "collector", proxy.serviceType)
}

func TestServiceProxy_NewServiceProxy_CoreServiceType(t *testing.T) {
	logger := setupTestLogger()
	proxy, err := NewServiceProxy("http://core-service:8082", "token", logger)
	assert.NoError(t, err)
	assert.Equal(t, "core", proxy.serviceType)
}

func TestAuthProxy_Success(t *testing.T) {
	backend := setupTestBackend()
	defer backend.Close()

	handler, err := AuthProxy(backend.URL)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/auth/test", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/auth/test")

	err = handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthProxy_InvalidURL(t *testing.T) {
	_, err := AuthProxy("://invalid-url")
	assert.Error(t, err)
}
