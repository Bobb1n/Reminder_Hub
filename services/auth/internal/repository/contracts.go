package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/auth/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindById(ctx context.Context, id uuid.UUID) (*models.User, error)
	IncrementVersion(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type BlacklistRepository interface {
	AddToken(ctx context.Context, tokenID string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
	CleanExpiredTokens(ctx context.Context) error
}
