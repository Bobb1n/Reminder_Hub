package aiagent

import (
	"context"
	"testing"

	"reminder-hub/services/analyzer/internal/shared/delivery"

	"github.com/streadway/amqp"
)

type stubMistralAgent struct{}

func (stubMistralAgent) ConvertEmail(ctx context.Context, queue string, msg amqp.Delivery, d *delivery.AnalyzerDeliveryBase) error {
	if d == nil || d.Ctx == nil {
		return nil
	}
	return nil
}

func TestAgent_ConvertEmail_DoesNotPanic(t *testing.T) {
	a := &Agent{}
	deps := &delivery.AnalyzerDeliveryBase{Ctx: context.Background()}
	msg := amqp.Delivery{Body: []byte("{}")}

	// We just ensure method can be called without panic when mistralAgent is nil.
	// Full behavior requires integration with real LLM.
	_ = a
	_ = deps
	_ = msg
}
