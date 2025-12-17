package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reminder-hub/pkg/logger"

	"auth/internal/domain/models"
)

// MockAuthUsecase - мок для AuthUsecase
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) SignUp(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthUsecase) SignIn(ctx context.Context, email, password string) (string, string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthUsecase) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthUsecase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	args := m.Called(ctx, accessToken, refreshToken)
	return args.Error(0)
}

func (m *MockAuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func setupTestRouter(handlers *AuthHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handlers.Register)
	router.POST("/auth/login", handlers.Login)
	router.POST("/auth/validate", handlers.ValidateToken)
	router.POST("/auth/logout", handlers.Logout)
	router.GET("/auth/me", handlers.GetCurrentUser)
	return router
}

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

func TestAuthHandlers_Register_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	mockUsecase.On("SignUp", mock.Anything, "test@example.com", "password123").Return(user, nil)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User created successfully", response["message"])
	assert.Equal(t, userID.String(), response["user_id"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Register_InvalidBody(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	body := map[string]string{
		"email": "invalid-email",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUsecase.AssertNotCalled(t, "SignUp")
}

func TestAuthHandlers_Register_EmailExists(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	mockUsecase.On("SignUp", mock.Anything, "existing@example.com", "password123").
		Return(nil, errors.New("email already exists"))

	body := map[string]string{
		"email":    "existing@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Email already registered", response["error"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Login_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	accessToken := "access-token-123"
	refreshToken := "refresh-token-456"

	mockUsecase.On("SignIn", mock.Anything, "test@example.com", "password123").
		Return(accessToken, refreshToken, nil)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, accessToken, response["access_token"])
	assert.Equal(t, refreshToken, response["refresh_token"])
	assert.Equal(t, "Bearer", response["token_type"])
	assert.Equal(t, float64(900), response["expires_in"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Login_InvalidCredentials(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	mockUsecase.On("SignIn", mock.Anything, "test@example.com", "wrongpassword").
		Return("", "", errors.New("invalid credentials"))

	body := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid credentials", response["error"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_ValidateToken_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	token := "valid-token-123"
	mockUsecase.On("ValidateToken", mock.Anything, token).Return(user, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["valid"])
	assert.Equal(t, userID.String(), response["user_id"])
	assert.Equal(t, "test@example.com", response["email"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_ValidateToken_NoHeader(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	req := httptest.NewRequest(http.MethodPost, "/auth/validate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUsecase.AssertNotCalled(t, "ValidateToken")
}

func TestAuthHandlers_ValidateToken_InvalidFormat(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	req := httptest.NewRequest(http.MethodPost, "/auth/validate", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUsecase.AssertNotCalled(t, "ValidateToken")
}

func TestAuthHandlers_ValidateToken_InvalidToken(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	token := "invalid-token"
	mockUsecase.On("ValidateToken", mock.Anything, token).Return(nil, errors.New("invalid token"))

	req := httptest.NewRequest(http.MethodPost, "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid token", response["error"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Logout_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	accessToken := "access-token-123"
	refreshToken := "refresh-token-456"

	mockUsecase.On("Logout", mock.Anything, accessToken, refreshToken).Return(nil)

	body := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Logged out successfully", response["message"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_GetCurrentUser_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	logger := setupTestLogger()
	handlers := NewAuthHandlers(mockUsecase, logger)
	router := setupTestRouter(handlers)

	userID := uuid.New()
	createdAt := time.Now()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: createdAt,
	}

	token := "valid-token-123"
	mockUsecase.On("ValidateToken", mock.Anything, token).Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, userID.String(), response["user_id"])
	assert.Equal(t, "test@example.com", response["email"])
	mockUsecase.AssertExpectations(t)
}

