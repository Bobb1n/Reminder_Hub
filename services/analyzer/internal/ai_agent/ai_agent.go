package aiagent

import (
	"context"
	"reminder-hub/services/analyzer/internal/ai_agent/mistral"
	"reminder-hub/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

type AiAgent interface {
	ConvertEmail(ctx context.Context, queue string, msg amqp.Delivery, dependencies *delivery.AnalyzerDeliveryBase) error
}

type Agent struct {
	mistralAgent *mistral.MistralAgent
}

func NewAgent(mistralAgent *mistral.MistralAgent) *Agent {
	return &Agent{mistralAgent: mistralAgent}
}

func (a *Agent) ConvertEmail(queue string, msg amqp.Delivery, dependencies *delivery.AnalyzerDeliveryBase) error {
	return a.mistralAgent.ConvertEmail(dependencies.Ctx, queue, msg, dependencies)
}
