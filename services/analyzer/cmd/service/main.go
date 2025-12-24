package main

import (
	ht "reminder-hub/pkg/http"
	"reminder-hub/pkg/rabbitmq"
	aiagent "reminder-hub/services/analyzer/internal/ai_agent"
	"reminder-hub/services/analyzer/internal/ai_agent/mistral"
	"reminder-hub/services/analyzer/internal/config"
	"reminder-hub/services/analyzer/internal/middleware/configurations"
	rc "reminder-hub/services/analyzer/internal/rabbitmq"
	"reminder-hub/services/analyzer/internal/server"
	"reminder-hub/services/analyzer/internal/server/echoserver"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Options(
			fx.Provide(
				config.InitConfig,
				ht.NewContext,
				echoserver.NewEchoServer,
				rabbitmq.NewRabbitMQConn,
				rabbitmq.NewPublisher,
				validator.New,
				mistral.NewMistralConn,
				aiagent.NewAgent,
			),
			fx.Invoke(server.RunServers),
			fx.Invoke(configurations.ConfigMiddlewares),
			fx.Invoke(rc.ConfigConsumers),
		),
	).Run()
}
