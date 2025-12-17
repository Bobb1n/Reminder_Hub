package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"reminder-hub/pkg/logger"

	"auth/internal/usecase"
)

const bearerPrefix = "Bearer"

type AuthHandlers struct {
	authUsecase usecase.AuthUsecase
	logger      *logger.CurrentLogger
}

func NewAuthHandlers(authUsecase usecase.AuthUsecase, log *logger.CurrentLogger) *AuthHandlers {
	return &AuthHandlers{
		authUsecase: authUsecase,
		logger:      log,
	}
}

// Register - регистрация нового пользователя
func (h *AuthHandlers) Register(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.authUsecase.SignUp(ctx, body.Email, body.Password)
	if err != nil {
		if err.Error() == "email already exists" {
			h.logger.Warn(ctx, "Registration attempt with existing email", "email", body.Email)
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		h.logger.Error(ctx, "Failed to create user", "email", body.Email, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.logger.Info(ctx, "User registered successfully", "user_id", user.ID, "email", body.Email)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

// Login - авторизация, выдача JWT токена
func (h *AuthHandlers) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	accessToken, refreshToken, err := h.authUsecase.SignIn(ctx, body.Email, body.Password)
	if err != nil {
		h.logger.Warn(ctx, "Login failed", "email", body.Email, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	h.logger.Info(ctx, "User logged in successfully", "email", body.Email)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
		"token_type":    "Bearer",
	})
}

// RefreshToken - обновление access token (внутренний endpoint, не регистрируется в публичных роутах) пока под вопросом
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
		return
	}

	ctx := c.Request.Context()
	accessToken, err := h.authUsecase.RefreshToken(ctx, body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   900,
		"token_type":   "Bearer",
	})
}

func (h *AuthHandlers) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format. Expected: Bearer <token>"})
		return
	}

	tokenString := parts[1]
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.authUsecase.ValidateToken(ctx, tokenString)
	if err != nil {
		h.logger.Warn(ctx, "Token validation failed", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	h.logger.Debug(ctx, "Token validated successfully", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (h *AuthHandlers) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
		return
	}

	accessToken := parts[1]

	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
		return
	}

	ctx := c.Request.Context()
	err := h.authUsecase.Logout(ctx, accessToken, body.RefreshToken)
	if err != nil {
		h.logger.Error(ctx, "Logout failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	h.logger.Info(ctx, "User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser - возвращает информацию о текущем пользователе
func (h *AuthHandlers) GetCurrentUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format. Expected: Bearer <token>"})
		return
	}

	tokenString := parts[1]
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.authUsecase.ValidateToken(ctx, tokenString)
	if err != nil {
		h.logger.Warn(ctx, "GetCurrentUser: token validation failed", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	h.logger.Debug(ctx, "GetCurrentUser: user info retrieved", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}
