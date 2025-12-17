package rabbit_configurations

import (
	"context"
	"testing"

	"reminder-hub/pkg/logger"
	rmq "reminder-hub/pkg/rabbitmq"
)

type nopLogger struct{}

func (nopLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (nopLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (nopLogger) Panic(ctx context.Context, msg string, fields ...any) {}
func (nopLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func newTestLogger() *logger.CurrentLogger { return logger.NewCurrentLogger(nopLogger{}) }

type stubPublisher struct{}

func (stubPublisher) PublishMessage(msg interface{}) error { return nil }
func (stubPublisher) IsPublished(msg interface{}) bool    { return false }

var _ rmq.IPublisher = (*stubPublisher)(nil)

func TestConfigConsumers_StartsConsumers(t *testing.T) {
	t.Skip("ConfigConsumers starts real AMQP consumers and requires live RabbitMQ; skipping in unit tests")
}
