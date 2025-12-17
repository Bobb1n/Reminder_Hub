package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"reminder-hub/pkg/logger"
)

// MockLogger - простой мок-логгер для тестов
type MockLogger struct{}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...any)  {}
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...any)   {}
func (m *MockLogger) Panic(ctx context.Context, msg string, fields ...any) {}
func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...any)  {}

func setupTestLogger() *logger.CurrentLogger {
	return logger.NewCurrentLogger(&MockLogger{})
}

func setupTestServer() *echo.Echo {
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/api/v1/test", func(c echo.Context) error {
		userID := c.Get("user_id")
		return c.JSON(http.StatusOK, map[string]interface{}{"user_id": userID})
	})
	return e
}

func TestAuthMiddleware_SkipHealthCheck(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "user_id": "test-user"})
	}))
	defer server.Close()

	e := setupTestServer()
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_SkipAuthRoutes(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"valid": true, "user_id": "test-user"})
	}))
	defer server.Close()

	e := setupTestServer()
	e.POST("/auth/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_NoAuthorizationHeader(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	e := setupTestServer()
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	e := setupTestServer()
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	logger := setupTestLogger()
	userID := "test-user-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   true,
			"user_id": userID,
			"email":   "test@example.com",
		})
	}))
	defer server.Close()

	e := setupTestServer()
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, userID, response["user_id"])
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
	}))
	defer server.Close()

	e := setupTestServer()
	e.Use(AuthMiddleware(server.URL, logger))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_AuthServiceUnavailable(t *testing.T) {
	logger := setupTestLogger()
	// Используем несуществующий URL для имитации недоступности сервиса
	e := setupTestServer()
	e.Use(AuthMiddleware("http://localhost:99999", logger))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Должен вернуть 401, так как сервис недоступен
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestValidateToken_Success(t *testing.T) {
	logger := setupTestLogger()
	userID := "test-user-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   true,
			"user_id": userID,
			"email":   "test@example.com",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := validateToken(server.URL, "test-token", logger, ctx)

	assert.NoError(t, err)
	assert.Equal(t, userID, result)
}

func TestValidateToken_InvalidResponse(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
		})
	}))
	defer server.Close()

	ctx := context.Background()
	_, err := validateToken(server.URL, "test-token", logger, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is not valid")
}

func TestValidateToken_EmptyUserID(t *testing.T) {
	logger := setupTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   true,
			"user_id": "",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	_, err := validateToken(server.URL, "test-token", logger, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is empty")
}

