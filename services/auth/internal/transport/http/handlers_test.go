package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/domain/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"go.uber.org/fx"
)

type mockAuthUsecase struct {
	mock.Mock
}

func (m *mockAuthUsecase) SignUp(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthUsecase) SignIn(ctx context.Context, email, password string) (string, string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockAuthUsecase) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *mockAuthUsecase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	args := m.Called(ctx, accessToken, refreshToken)
	return args.Error(0)
}

func (m *mockAuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func setupTestHandlers() (*AuthHandlers, *mockAuthUsecase) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockAuthUsecase)
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)
	handlers := NewAuthHandlers(mockUsecase, logger)
	return handlers, mockUsecase
}

type simpleLifecycle struct{}

func (s *simpleLifecycle) Append(hook fx.Hook) {}

func TestAuthHandlers_Register_Success(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	userID := uuid.New()
	mockUsecase.On("SignUp", mock.Anything, "test@example.com", "password123").
		Return(&models.User{ID: userID, Email: "test@example.com"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", 
		bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Register_InvalidBody(t *testing.T) {
	handlers, _ := setupTestHandlers()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", 
		bytes.NewBufferString(`{"email":"invalid"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlers_Register_EmailExists(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	mockUsecase.On("SignUp", mock.Anything, "test@example.com", "password123").
		Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", 
		bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Register(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Login_Success(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	mockUsecase.On("SignIn", mock.Anything, "test@example.com", "password123").
		Return("access_token", "refresh_token", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", 
		bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "access_token", response["access_token"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_Login_InvalidCredentials(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	mockUsecase.On("SignIn", mock.Anything, "test@example.com", "wrong").
		Return("", "", assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", 
		bytes.NewBufferString(`{"email":"test@example.com","password":"wrong"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_ValidateToken_Success(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	userID := uuid.New()
	mockUsecase.On("ValidateToken", mock.Anything, "valid_token").
		Return(&models.User{ID: userID, Email: "test@example.com"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
	c.Request.Header.Set("Authorization", "Bearer valid_token")

	handlers.ValidateToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["valid"])
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_ValidateToken_NoHeader(t *testing.T) {
	handlers, _ := setupTestHandlers()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/validate", nil)

	handlers.ValidateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlers_ValidateToken_InvalidFormat(t *testing.T) {
	handlers, _ := setupTestHandlers()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat")

	handlers.ValidateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlers_Logout_Success(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	mockUsecase.On("Logout", mock.Anything, "access_token", "refresh_token").
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", 
		bytes.NewBufferString(`{"refresh_token":"refresh_token"}`))
	c.Request.Header.Set("Authorization", "Bearer access_token")
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandlers_GetCurrentUser_Success(t *testing.T) {
	handlers, mockUsecase := setupTestHandlers()
	
	userID := uuid.New()
	mockUsecase.On("ValidateToken", mock.Anything, "valid_token").
		Return(&models.User{ID: userID, Email: "test@example.com"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Request.Header.Set("Authorization", "Bearer valid_token")

	handlers.GetCurrentUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, userID.String(), response["user_id"])
	mockUsecase.AssertExpectations(t)
}
