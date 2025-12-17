package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry_Success(t *testing.T) {
	cfg := DefaultRetryConfig()
	attempts := 0
	
	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_RetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()
	attempts := 0
	retryableErr := errors.New("timeout error")
	
	err := Retry(context.Background(), cfg, func() error {
		attempts++
		if attempts < 2 {
			return retryableErr
		}
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_MaxAttempts(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 2
	attempts := 0
	
	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return errors.New("persistent error")
	})
	
	assert.Error(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_ContextCancellation(t *testing.T) {
	cfg := DefaultRetryConfig()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу
	
	err := Retry(ctx, cfg, func() error {
		return errors.New("error")
	})
	
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestIsRetryableError_Retryable(t *testing.T) {
	assert.True(t, IsRetryableError(errors.New("timeout error")))
	assert.True(t, IsRetryableError(errors.New("connection refused")))
	assert.True(t, IsRetryableError(errors.New("rate limit exceeded")))
}

func TestIsRetryableError_NotRetryable(t *testing.T) {
	assert.False(t, IsRetryableError(nil))
	assert.False(t, IsRetryableError(errors.New("invalid input")))
	assert.False(t, IsRetryableError(context.Canceled))
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 10*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
}

