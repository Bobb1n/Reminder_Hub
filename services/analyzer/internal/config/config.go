package config

import (
	"flag"
	"fmt"
	"log"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/logger/zaplogger"
	"reminder-hub/pkg/rabbitmq"
	"reminder-hub/services/analyzer/internal/ai_agent/mistral"
	"reminder-hub/services/analyzer/internal/server/echoserver"
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

	adapter := zaplogger.NewLoggerAdapter(fx, cfg.Environment)

	log.Printf("Config WAS PARSED. \n\n THERE IS SOME VALUES: Config:%v\n EchoConfig:%v\n RabbitConfig: %v\n", cfg, cfg.Echo, cfg.Rabbitmq)

	return cfg, logger.NewCurrentLogger(adapter), cfg.Echo, cfg.Rabbitmq, cfg.MistralConfig, nil
}

func GetMicroserviceName(serviceName string) string {
	return fmt.Sprintf("%s", strings.ToUpper(serviceName))
}
