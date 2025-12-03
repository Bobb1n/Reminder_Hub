package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/auth/internal/domain/models"
)

type UserRepo struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	id := uuid.New()
	now := time.Now()
	query, args, err := r.builder.Insert("users").
		Columns("id", "email", "password_hash", "version", "created_at", "updated_at").
		Values(id, email, passwordHash, 1, now, now).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("email already exists")
		}

		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &models.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := r.builder.Select("id", "email", "password_hash", "version", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("find by email: %w", err)
	}
	return r.findUserByQuery(ctx, query, args, "find by email")
}

func (r *UserRepo) FindById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query, args, err := r.builder.Select("id", "email", "password_hash", "version", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return r.findUserByQuery(ctx, query, args, "find by id")
}

func (r *UserRepo) findUserByQuery(
	ctx context.Context,
	query string,
	args []any,
	operation string,
) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return &user, nil
}

func (r *UserRepo) IncrementVersion(ctx context.Context, userID uuid.UUID) error {
	query, args, err := r.builder.Update("users").
		Set("version", squirrel.Expr("version + 1")).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("increment version: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("increment version: %w", err)
	}

	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query, args, err := r.builder.Update("users").
		Set("password_hash", passwordHash).
		Set("version", squirrel.Expr("version + 1")).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
