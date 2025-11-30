package main

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Options(
			fx.Provide(
				config.InitConfig,
				logger.InitLogger,
				http.NewContext,
				echoserver.NewEchoServer,
				httpclient.NewHttpClient,
				repositories.NewPostgresProductRepository,
				rabbitmq.NewRabbitMQConn,
				rabbitmq.NewPublisher,
				validator.New,
			),
			fx.Invoke(server.RunServers),
			fx.Invoke(configurations.ConfigMiddlewares),
			fx.Invoke(configurations.ConfigSwagger),
			fx.Invoke(mappings.ConfigureMappings),
			fx.Invoke(configurations.ConfigEndpoints),
			fx.Invoke(configurations.ConfigProductsMediator),
		),
	).Run()
}
