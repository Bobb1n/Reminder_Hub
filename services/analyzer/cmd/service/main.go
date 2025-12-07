package main

import (
	ht "yfp/internal/http"
	"yfp/internal/rabbitmq"
	aiagent "yfp/services/analyzer/internal/ai_agent"
	"yfp/services/analyzer/internal/ai_agent/mistral"
	"yfp/services/analyzer/internal/config"
	"yfp/services/analyzer/internal/middleware/configurations"
	rc "yfp/services/analyzer/internal/rabbitmq"
	"yfp/services/analyzer/internal/server"
	"yfp/services/analyzer/internal/server/echoserver"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Options(
			fx.Provide(
				config.InitConfig,        //✅
				ht.NewContext,            //✅
				echoserver.NewEchoServer, //✅
				rabbitmq.NewRabbitMQConn, //✅
				rabbitmq.NewPublisher,    //✅
				validator.New,            //✅
				mistral.NewMistralConn,
				aiagent.NewAgent,
			),
			fx.Invoke(server.RunServers),                //✅
			fx.Invoke(configurations.ConfigMiddlewares), //✅
			fx.Invoke(rc.ConfigConsumers),               //✅
		),
	).Run()
}
