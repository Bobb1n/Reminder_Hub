package main

import (
	ht "yfp/internal/http"
	"yfp/internal/logger"
	"yfp/internal/rabbitmq"
	aiagent "yfp/services/analyzer/internal/ai_agent"
	"yfp/services/analyzer/internal/config"
	"yfp/services/analyzer/internal/middleware/configurations"
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
				logger.NewCurrentLogger,  //✅
				ht.NewContext,            //✅
				echoserver.NewEchoServer, //✅
				rabbitmq.NewRabbitMQConn, //✅
				rabbitmq.NewPublisher,    //✅
				rabbitmq.NewConsumer,     //✅
				validator.New,            //✅
				aiagent.NewAgent,         //❌
			),
			fx.Invoke(server.RunServers),                //✅
			fx.Invoke(configurations.ConfigMiddlewares), //✅
		),
	).Run()
}
