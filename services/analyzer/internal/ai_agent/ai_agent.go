package aiagent

import (
	"context"
	"yfp/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

func NewAgent(ctx context.Context) (llm, error) {

}

func ConvertEmail(queue string, msg amqp.Delivery, dependencies *delivery.AnalyzerDeliveryBase) error {

}
