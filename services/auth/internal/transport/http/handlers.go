package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/auth/internal/usecase"
)

const bearerPrefix = "Bearer"

type AuthHandlers struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandlers(authUsecase usecase.AuthUsecase) *AuthHandlers {
	return &AuthHandlers{
		authUsecase: authUsecase,
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
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}
