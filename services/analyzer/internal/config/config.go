package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"reminder-hub/pkg/rabbitmq"
	"reminder-hub/services/analyzer/internal/ai_agent/mistral"
	"reminder-hub/services/analyzer/internal/server/echoserver"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var configPath string

type Config struct {
	Environment   string `env:"ENV" env-default:"development"`
	ServiceName   string `env:"SERVICE_NAME" env-default:"analyzer"`
	Logger        *logger.CurrentLogger
	Rabbitmq      *rabbitmq.RabbitMQConfig `env-prefix:"RABBITMQ_"`
	Echo          *echoserver.EchoConfig   `env-prefix:"ECHO_"`
	MistralConfig *mistral.MistralConfig   `env-prefix:"MISTRAL_"`
}

func init() {
	flag.StringVar(&configPath, "config", "", "products write microservice config path")
}

func InitConfig(fx fx.Lifecycle) (*Config, *logger.CurrentLogger, *echoserver.EchoConfig, *rabbitmq.RabbitMQConfig, *mistral.MistralConfig, error) {

	_ = godotenv.Load(".env")
	cfg := &Config{
		Rabbitmq:      &rabbitmq.RabbitMQConfig{},
		Echo:          &echoserver.EchoConfig{},
		MistralConfig: &mistral.MistralConfig{},
	}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse config %w", err)
	}

	// Явно парсим RabbitMQ конфигурацию с fallback на переменные окружения
	cfg.Rabbitmq = parseRabbitMQConfig()

	adapter := zaplogger.NewLoggerAdapter(fx, cfg.Environment)

	log.Printf("Config WAS PARSED. \n\n THERE IS SOME VALUES: Config:%v\n EchoConfig:%v\n RabbitConfig: Host=%s, Port=%d, User=%s\n", cfg, cfg.Echo, cfg.Rabbitmq.Host, cfg.Rabbitmq.Port, cfg.Rabbitmq.User)

	return cfg, logger.NewCurrentLogger(adapter), cfg.Echo, cfg.Rabbitmq, cfg.MistralConfig, nil
}

// parseRabbitMQConfig парсит конфигурацию RabbitMQ из переменных окружения с fallback на значения по умолчанию
func parseRabbitMQConfig() *rabbitmq.RabbitMQConfig {
	port := 5672
	if portStr := os.Getenv("RABBITMQ_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &rabbitmq.RabbitMQConfig{
		Host:         getEnvOrDefault("RABBITMQ_HOST", "localhost"),
		Port:         port,
		User:         getEnvOrDefault("RABBITMQ_USER", "guest"),
		Password:     getEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
		ExchangeName: getEnvOrDefault("RABBITMQ_EXCHANGE", "donotmatter"),
		Kind:         getEnvOrDefault("RABBITMQ_KIND", "topic"),
	}
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetMicroserviceName(serviceName string) string {
	return fmt.Sprintf("%s", strings.ToUpper(serviceName))
}
