package configurations

import (
	"context"
	"yfp/internal/logger"
	"yfp/internal/models"
	"yfp/internal/rabbitmq"
	aiagent "yfp/services/analyzer/internal/ai_agent"
	"yfp/services/analyzer/internal/config"
	"yfp/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

func ConfigConsumers(
	ctx context.Context,
	log logger.CurrentLogger,
	connRabbitmq *amqp.Connection,
	rabbitmqPublisher rabbitmq.IPublisher,
	cfg *config.Config) error {

	inventoryDeliveryBase := delivery.AnalyzerDeliveryBase{
		Log:               log,
		Cfg:               cfg,
		ConnRabbitmq:      connRabbitmq,
		RabbitmqPublisher: rabbitmqPublisher,
		Ctx:               ctx,
	}

	createProductConsumer := rabbitmq.NewConsumer[*delivery.AnalyzerDeliveryBase](ctx, cfg.Rabbitmq, connRabbitmq, log, aiagent.ConvertEmail)

	go func() {
		err := createProductConsumer.ConsumeMessage(models.RawEmails{}, &inventoryDeliveryBase)
		if err != nil {
			log.Error(ctx, "ConfigConsumers error in func ConsumeMessage: ", err)
		}
	}()

	return nil
}
