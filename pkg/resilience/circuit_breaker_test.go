package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)
	
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Second)
	
	// Первая ошибка
	cb.Execute(context.Background(), func() error {
		return errors.New("error 1")
	})
	assert.Equal(t, StateClosed, cb.State())
	
	// Вторая ошибка - открываем circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error 2")
	})
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_OpenState_Rejects(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond)
	
	// Открываем circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	
	// Пытаемся выполнить - должно быть отклонено
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestCircuitBreaker_HalfOpen_Recovery(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	
	// Открываем circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	assert.Equal(t, StateOpen, cb.State())
	
	// Ждем reset timeout
	time.Sleep(60 * time.Millisecond)
	
	// Успешное выполнение должно закрыть circuit
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)
	
	// Две ошибки
	cb.Execute(context.Background(), func() error { return errors.New("error") })
	cb.Execute(context.Background(), func() error { return errors.New("error") })
	
	// Успешное выполнение сбрасывает счетчик
	cb.Execute(context.Background(), func() error { return nil })
	
	assert.Equal(t, StateClosed, cb.State())
}

