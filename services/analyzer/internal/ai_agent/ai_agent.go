package aiagent

import (
	"reminder-hub/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

type AiAgent interface {
	ConvertEmail(queue string, msg amqp.Delivery, dependencies *delivery.AnalyzerDeliveryBase) error
}

type Agent struct {
	ai AiAgent
}

func NewAgent(ai AiAgent) *Agent {
	return &Agent{ai: ai}
}

func (a *Agent) ConvertEmail(queue string, msg amqp.Delivery, dependencies *delivery.AnalyzerDeliveryBase) error {
	return a.ConvertEmail(queue, msg, dependencies)
}
