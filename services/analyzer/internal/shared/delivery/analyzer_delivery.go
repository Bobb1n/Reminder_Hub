package delivery

import (
	"context"
	"yfp/internal/logger"
	"yfp/internal/rabbitmq"
	"yfp/services/analyzer/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

type AnalyzerDeliveryBase struct {
	Log               logger.CurrentLogger
	Cfg               *config.Config
	RabbitmqPublisher rabbitmq.IPublisher
	ConnRabbitmq      *amqp.Connection
	Echo              *echo.Echo
	Ctx               context.Context'
	LLM 
}
