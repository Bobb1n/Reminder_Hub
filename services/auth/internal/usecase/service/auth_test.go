package service

import (
	"context"
	"testing"
	"time"

	"auth/internal/domain/models"
	"auth/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRepo struct {
	user *models.User
}

func (s *stubUserRepo) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	return &models.User{ID: s.user.ID, Email: email, PasswordHash: passwordHash}, nil
}
func (s *stubUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) FindById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) IncrementVersion(ctx context.Context, userID uuid.UUID) error { return nil }
func (s *stubUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	return nil
}

type blacklistedToken struct {
	id  string
	exp time.Time
}

type stubBlacklistRepo struct {
	tokens []blacklistedToken
}

func (s *stubBlacklistRepo) IsTokenBlacklisted(ctx context.Context, id string) (bool, error) {
	for _, t := range s.tokens {
		if t.id == id {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubBlacklistRepo) AddToken(ctx context.Context, id string, exp time.Time) error {
	s.tokens = append(s.tokens, blacklistedToken{id: id, exp: exp})
	return nil
}

func (s *stubBlacklistRepo) CleanExpiredTokens(ctx context.Context) error { return nil }

var _ repository.UserRepository = (*stubUserRepo)(nil)
var _ repository.BlacklistRepository = (*stubBlacklistRepo)(nil)

func TestAuthService_SignUpAndSignIn(t *testing.T) {
	ctx := context.Background()

	// prepare user with known password hash
	rawPassword := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword error: %v", err)
	}

	user := &models.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: string(hash)}
	us := &stubUserRepo{user: user}
	bs := &stubBlacklistRepo{}
	service := NewAuthService(us, bs, "secret-key")

	// SignIn should succeed with correct credentials
	access, refresh, err := service.SignIn(ctx, user.Email, rawPassword)
	if err != nil {
		t.Fatalf("SignIn error=%v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected non-empty tokens, got access=%q refresh=%q", access, refresh)
	}

	// RefreshToken should return new access token
	newAccess, err := service.RefreshToken(ctx, refresh)
	if err != nil {
		t.Fatalf("RefreshToken error=%v", err)
	}
	if newAccess == "" {
		t.Fatal("expected non-empty new access token")
	}

	// ValidateToken should return a user for valid access token
	validatedUser, err := service.ValidateToken(ctx, access)
	if err != nil {
		t.Fatalf("ValidateToken error=%v", err)
	}
	if validatedUser.Email != user.Email {
		t.Fatalf("ValidateToken returned user email %q, want %q", validatedUser.Email, user.Email)
	}

	// Logout should blacklist both tokens
	if err := service.Logout(ctx, access, refresh); err != nil {
		t.Fatalf("Logout error=%v", err)
	}
}
