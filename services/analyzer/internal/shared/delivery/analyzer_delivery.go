package delivery

import (
	"context"
	"yfp/internal/logger"
	"yfp/internal/rabbitmq"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

type AnalyzerDeliveryBase struct {
	Log               logger.CurrentLogger
	RabbitmqPublisher rabbitmq.IPublisher
	ConnRabbitmq      *amqp.Connection
	Echo              *echo.Echo
	Ctx               context.Context
}
