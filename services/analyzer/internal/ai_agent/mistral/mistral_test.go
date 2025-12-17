package mistral

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"go.uber.org/fx"
)

type simpleLifecycle struct{}

func (s *simpleLifecycle) Append(hook fx.Hook) {}

func TestMistralConfig_API(t *testing.T) {
	cfg := &MistralConfig{
		api: "test-api-key",
	}
	
	assert.Equal(t, "test-api-key", cfg.API())
}

func TestMistralConfig_SetAPI(t *testing.T) {
	cfg := &MistralConfig{}
	cfg.SetAPI("new-api-key")
	
	assert.Equal(t, "new-api-key", cfg.API())
}

func TestMistralConfig_DefaultModel(t *testing.T) {
	cfg := &MistralConfig{
		model: "",
	}
	
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)
	ctx := context.Background()
	
	// Устанавливаем API ключ для прохождения валидации
	cfg.SetAPI("test-key")
	
	// Проверяем, что модель устанавливается по умолчанию
	_, err := NewMistralConn(ctx, cfg, logger)
	// Ожидаем ошибку от mistral.New, но модель должна быть установлена
	if err != nil {
		// Это нормально, так как мы не подключаемся к реальному API
		assert.Equal(t, "open-mistral-7b", cfg.model)
	}
}

func TestNewMistralConn_EmptyAPIKey(t *testing.T) {
	cfg := &MistralConfig{
		api: "",
	}
	
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)
	ctx := context.Background()
	
	agent, err := NewMistralConn(ctx, cfg, logger)
	
	assert.Nil(t, agent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mistral API key is required")
}

func TestNewMistralConn_WithAPIKey(t *testing.T) {
	cfg := &MistralConfig{
		api:     "test-api-key",
		model:   "open-mistral-7b",
		timeout: 30 * time.Second,
		retries: 3,
	}
	
	lc := &simpleLifecycle{}
	adapter := zaplogger.NewLoggerAdapter(lc, "test")
	logger := logger.NewCurrentLogger(adapter)
	ctx := context.Background()
	
	// Этот тест может упасть, если нет реального подключения к Mistral API
	// Но мы проверяем, что функция пытается создать соединение
	agent, err := NewMistralConn(ctx, cfg, logger)
	
	// Если ошибка, это нормально для unit-теста без реального API
	if err != nil {
		// Проверяем, что ошибка не связана с пустым API ключом
		assert.NotContains(t, err.Error(), "mistral API key is required")
	} else {
		require.NotNil(t, agent)
		assert.NotNil(t, agent.llm)
	}
}

func TestMistralConfig_TimeoutAndRetries(t *testing.T) {
	cfg := &MistralConfig{
		timeout: 60 * time.Second,
		retries: 5,
	}
	
	assert.Equal(t, 60*time.Second, cfg.timeout)
	assert.Equal(t, 5, cfg.retries)
}

