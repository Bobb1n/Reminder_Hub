package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BlacklistRepo struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewBlacklistRepo(db *pgxpool.Pool) *BlacklistRepo {
	return &BlacklistRepo{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *BlacklistRepo) AddToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	// ON CONFLICT DO NOTHING  позволяет безопасно вызывать logout несколько раз с теми же токенами
	query := `
		INSERT INTO blacklisted_tokens (token_id, expires_at, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, tokenID, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("add token to blacklist: %w", err)
	}

	return nil
}

func (r *BlacklistRepo) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	query, args, err := r.builder.Select("COUNT(*)").
		From("blacklisted_tokens").
		Where(squirrel.And{
			squirrel.Eq{"token_id": tokenID},
			squirrel.Gt{"expires_at": time.Now()},
		}).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	var count int
	err = r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	return count > 0, nil
}

func (r *BlacklistRepo) CleanExpiredTokens(ctx context.Context) error {
	query, args, err := r.builder.Delete("blacklisted_tokens").
		Where(squirrel.Lt{"expires_at": time.Now()}).
		ToSql()

	if err != nil {
		return fmt.Errorf("clean expired tokens: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("clean expired tokens: %w", err)
	}

	return nil
}
