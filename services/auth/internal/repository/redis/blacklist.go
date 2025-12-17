package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BlacklistRepo struct {
	client *redis.Client
	prefix string
}

func NewBlacklistRepo(client *redis.Client) *BlacklistRepo {
	return &BlacklistRepo{
		client: client,
		prefix: "blacklist:token:",
	}
}

func (r *BlacklistRepo) AddToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	key := r.prefix + tokenID

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	err := r.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("add token to blacklist: %w", err)
	}

	return nil
}

func (r *BlacklistRepo) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := r.prefix + tokenID

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	return count > 0, nil
}

func (r *BlacklistRepo) CleanExpiredTokens(ctx context.Context) error {
	return nil
}
