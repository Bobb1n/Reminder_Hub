package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"auth/internal/domain/models"
	"auth/internal/repository"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type AuthService struct {
	userRepo      repository.UserRepository
	blacklistRepo repository.BlacklistRepository
	jwtSecret     string
}

func NewAuthService(
	userRepo repository.UserRepository,
	blacklistRepo repository.BlacklistRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		jwtSecret:     jwtSecret,
	}
}

func (s *AuthService) SignUp(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, email, string(hash))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) SignIn(
	ctx context.Context,
	email, password string,
) (accessToken, refreshToken string, err error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err = s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.generateRefreshToken(user.ID, user.Version)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) generateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"type":    "access",
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateRefreshToken(userID uuid.UUID, version int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"type":    "refresh",
		"version": version,
		"exp":     time.Now().Add(refreshTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (*AuthService) getTokenID(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	tokenID := s.getTokenID(refreshTokenString)
	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return "", fmt.Errorf("check blacklist: %w", err)
	}
	if isBlacklisted {
		return "", errors.New("token has been revoked")
	}

	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", errors.New("invalid token type: expected refresh token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id not found in token")
	}

	tokenVersion, ok := claims["version"].(float64)
	if !ok {
		return "", errors.New("version not found in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user_id: %w", err)
	}

	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	if user.Version != int(tokenVersion) {
		return "", errors.New("token version mismatch - please login again")
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	return accessToken, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	tokenID := s.getTokenID(tokenString)
	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id not found in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	accessTokenID := s.getTokenID(accessToken)
	refreshTokenID := s.getTokenID(refreshToken)

	// Извлекаем реальное время истечения из токенов
	accessExpiresAt, err := s.getTokenExpiration(accessToken)
	if err != nil {
		accessExpiresAt = time.Now().Add(accessTokenTTL)
	}

	refreshExpiresAt, err := s.getTokenExpiration(refreshToken)
	if err != nil {
		refreshExpiresAt = time.Now().Add(refreshTokenTTL)
	}

	if err := s.blacklistRepo.AddToken(ctx, accessTokenID, accessExpiresAt); err != nil {
		return fmt.Errorf("blacklist access token: %w", err)
	}

	if err := s.blacklistRepo.AddToken(ctx, refreshTokenID, refreshExpiresAt); err != nil {
		return fmt.Errorf("blacklist refresh token: %w", err)
	}

	return nil
}

// getTokenExpiration извлекает время истечения из JWT токена
func (s *AuthService) getTokenExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, errors.New("exp claim not found")
	}

	return time.Unix(int64(exp), 0), nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("invalid old password")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.userRepo.UpdatePassword(ctx, userID, string(newHash))
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
