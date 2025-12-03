package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/auth/internal/domain/models"
)

type AuthUsecase interface {
	SignUp(ctx context.Context, email, password string) (*models.User, error)
	SignIn(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshTokenString string) (accessToken string, err error)
	ValidateToken(ctx context.Context, tokenString string) (*models.User, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
