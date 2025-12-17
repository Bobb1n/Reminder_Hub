package resilience

import (
	"context"
	"errors"
	"time"
)

// RetryConfig настройки для retry механизма
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig возвращает конфигурацию по умолчанию
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// Retry выполняет функцию с повторными попытками при ошибках
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Проверяем контекст перед каждой попыткой
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Не делаем задержку после последней попытки
		if attempt < cfg.MaxAttempts-1 {
			// Экспоненциальная задержка с ограничением
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return lastErr
}

// IsRetryableError проверяет, можно ли повторить операцию при данной ошибке
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Сетевые ошибки и таймауты - можно повторить
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false // Контекст отменен - не повторяем
	}

	// Ошибки типа "connection refused", "timeout" и т.д. - можно повторить
	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"rate limit",
		"too many requests",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

