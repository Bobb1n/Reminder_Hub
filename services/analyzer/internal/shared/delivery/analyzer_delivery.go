package delivery

import (
	"context"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/rabbitmq"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

type AnalyzerDeliveryBase struct {
	Log               *logger.CurrentLogger
	RabbitmqPublisher rabbitmq.IPublisher
	ConnRabbitmq      *amqp.Connection
	Echo              *echo.Echo
	Ctx               context.Context
}
