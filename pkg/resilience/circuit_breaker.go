package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State состояние circuit breaker
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker реализует паттерн circuit breaker
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	state         State
	failureCount  int
	lastFailTime  time.Time
	mu            sync.RWMutex
}

// NewCircuitBreaker создает новый circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute выполняет функцию через circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	switch state {
	case StateOpen:
		// Проверяем, можно ли перейти в half-open
		cb.mu.Lock()
		if time.Since(cb.lastFailTime) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.failureCount = 0
			state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is open")
		}
		cb.mu.Unlock()

	case StateHalfOpen:
		// Пробуем выполнить операцию
		break

	case StateClosed:
		// Нормальная работа
		break
	}

	// Выполняем функцию
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Успешное выполнение - сбрасываем счетчик и закрываем circuit
	cb.failureCount = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}

	return nil
}

// State возвращает текущее состояние
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

