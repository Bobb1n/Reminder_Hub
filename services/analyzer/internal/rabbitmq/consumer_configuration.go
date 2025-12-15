package rabbit_configurations

import (
	"context"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/models"
	rmq "reminder-hub/pkg/rabbitmq"
	aiagent "reminder-hub/services/analyzer/internal/ai_agent"
	"reminder-hub/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

const numberOfConsumers = 4

func ConfigConsumers(
	ctx context.Context,
	log *logger.CurrentLogger,
	connRabbitmq *amqp.Connection,
	rabbitmqPublisher rmq.IPublisher,
	aiagent *aiagent.Agent,
	rabbitmq *rmq.RabbitMQConfig,
) error {

	inventoryDeliveryBase := delivery.AnalyzerDeliveryBase{
		Log:               log,
		ConnRabbitmq:      connRabbitmq,
		RabbitmqPublisher: rabbitmqPublisher,
		Ctx:               ctx,
	}

	createProductConsumer := rmq.NewConsumer[*delivery.AnalyzerDeliveryBase](ctx, rabbitmq, connRabbitmq, log, aiagent.ConvertEmail)
	for i := 0; i < numberOfConsumers; i++ {
		go func() {
			err := createProductConsumer.ConsumeMessage(models.RawEmails{}, &inventoryDeliveryBase)
			if err != nil {
				log.Error(ctx, "ConfigConsumers error in func ConsumeMessage: ", err)
			}
		}()
	}

	return nil
}
