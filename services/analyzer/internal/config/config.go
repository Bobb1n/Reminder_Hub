package config

import (
	"flag"
	"fmt"
	"strings"
	"yfp/internal/logger"
	"yfp/internal/rabbitmq"
	"yfp/services/analyzer/internal/ai_agent/mistral"
	"yfp/services/analyzer/internal/server/echoserver"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
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

func InitConfig() (*Config, *logger.CurrentLogger, *echoserver.EchoConfig, *rabbitmq.RabbitMQConfig, *mistral.MistralConfig, error) {

	_ = godotenv.Load(".env")
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse config %w", err)
	}

	return cfg, cfg.Logger, cfg.Echo, cfg.Rabbitmq, cfg.MistralConfig, nil
}

func GetMicroserviceName(serviceName string) string {
	return fmt.Sprintf("%s", strings.ToUpper(serviceName))
}
