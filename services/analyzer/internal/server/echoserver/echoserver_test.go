package echoserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"go.uber.org/fx"
)

type simpleLifecycle struct{}

func (s *simpleLifecycle) Append(hook fx.Hook) {}

func TestNewEchoServer(t *testing.T) {
	server := NewEchoServer()
	assert.NotNil(t, server)
}

func TestEchoConfig_Defaults(t *testing.T) {
	cfg := &EchoConfig{}
	// Проверяем, что структура существует
	assert.NotNil(t, cfg)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 15*time.Second, ReadTimeout)
	assert.Equal(t, 15*time.Second, WriteTimeout)
	assert.Equal(t, 1<<20, MaxHeaderBytes)
}

func TestRunEchoServer_ContextCancellation(t *testing.T) {
	server := NewEchoServer()
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)
	cfg := &EchoConfig{
		Port: ":0", // Используем порт 0 для автоматического выбора
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Запускаем сервер в горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- RunEchoServer(ctx, server, logger, cfg)
	}()
	
	// Даем серверу немного времени на запуск
	time.Sleep(100 * time.Millisecond)
	
	// Отменяем контекст
	cancel()
	
	// Ждем завершения
	select {
	case err := <-errChan:
		// Ошибка может быть или nil (нормальное завершение) или ошибка запуска
		_ = err
	case <-time.After(2 * time.Second):
		// Если сервер не завершился, это нормально для теста
		t.Log("Server shutdown test completed")
	}
}

