package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlacklistRepo_AddToken_Success(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Используем отдельную БД для тестов
	})
	defer client.Close()

	ctx := context.Background()
	repo := NewBlacklistRepo(client)

	// Очищаем тестовую БД
	client.FlushDB(ctx)

	tokenID := "test-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)

	err := repo.AddToken(ctx, tokenID, expiresAt)
	require.NoError(t, err)

	// Проверяем, что токен добавлен
	exists, err := repo.IsTokenBlacklisted(ctx, tokenID)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestBlacklistRepo_AddToken_ExpiredToken(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	ctx := context.Background()
	repo := NewBlacklistRepo(client)

	client.FlushDB(ctx)

	tokenID := "expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Уже истек

	err := repo.AddToken(ctx, tokenID, expiresAt)
	require.NoError(t, err) // Не должна возвращать ошибку для истекших токенов

	// Проверяем, что токен НЕ добавлен (так как уже истек)
	exists, err := repo.IsTokenBlacklisted(ctx, tokenID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBlacklistRepo_IsTokenBlacklisted_NotExists(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	ctx := context.Background()
	repo := NewBlacklistRepo(client)

	client.FlushDB(ctx)

	exists, err := repo.IsTokenBlacklisted(ctx, "non-existent-token")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBlacklistRepo_IsTokenBlacklisted_Exists(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	ctx := context.Background()
	repo := NewBlacklistRepo(client)

	client.FlushDB(ctx)

	tokenID := "existing-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	err := repo.AddToken(ctx, tokenID, expiresAt)
	require.NoError(t, err)

	exists, err := repo.IsTokenBlacklisted(ctx, tokenID)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestBlacklistRepo_CleanExpiredTokens(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	ctx := context.Background()
	repo := NewBlacklistRepo(client)

	client.FlushDB(ctx)

	// CleanExpiredTokens всегда возвращает nil (токены удаляются автоматически по TTL)
	err := repo.CleanExpiredTokens(ctx)
	require.NoError(t, err)
}

func TestNewBlacklistRepo(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	repo := NewBlacklistRepo(client)
	require.NotNil(t, repo)
	assert.Equal(t, "blacklist:token:", repo.prefix)
}

